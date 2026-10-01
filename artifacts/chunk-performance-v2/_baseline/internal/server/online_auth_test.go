package server

import (
	"atlas/internal/hub"
	"atlas/internal/proto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSessionHashUsesSignedJavaSHA1(t *testing.T) {
	secret := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	if got := sessionHash(secret, []byte("hello")); got != "-649fd9206783484a165a1ca940514ed055e93823" {
		t.Fatalf("signed hash = %s", got)
	}
	if got := sessionHash(secret, []byte("abc")); got != "194e92ab3fef38270c5a5afedbd47b1c06a7d333" {
		t.Fatalf("positive hash = %s", got)
	}
}

func TestOnlineLoginVerifiesSessionAndUsesOfficialProfile(t *testing.T) {
	const uuidHex = "8072fec8688f400699ea4e74deeec5d7"
	query := make(chan string, 1)
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query <- r.URL.Query().Get("username") + ":" + r.URL.Query().Get("serverId")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"id":"%s","name":"Yun_mo_an","properties":[{"name":"textures","value":"texture-data","signature":"signed-texture"}]}`, uuidHex)
	}))
	defer httpServer.Close()
	s, client, done := onlineLoginPipe(t, httpServer.URL, "Yun_mo_an")
	defer client.Conn.Close()
	secret, publicDER := answerEncryptionRequest(t, client, false)
	if id, payload, err := client.ReadPacket(); err != nil || id != proto.SLoginSetCompress {
		t.Fatalf("encrypted compression packet: id=%d payload=%x err=%v", id, payload, err)
	} else {
		threshold, err := proto.NewReader(payload).VarInt()
		if err != nil || threshold != 256 {
			t.Fatalf("compression threshold = %d, %v", threshold, err)
		}
		client.EnableCompression(threshold)
	}
	id, payload, err := client.ReadPacket()
	if err != nil || id != proto.SLoginSuccess {
		t.Fatalf("encrypted Login Success: id=%d err=%v", id, err)
	}
	r := proto.NewReader(payload)
	uuid, err := r.UUID()
	if err != nil || formatUUID(uuid) != "8072fec8-688f-4006-99ea-4e74deeec5d7" {
		t.Fatalf("profile UUID = %s, %v", formatUUID(uuid), err)
	}
	name, _ := r.String(16)
	count, _ := r.VarInt()
	key, _ := r.String(0)
	value, _ := r.String(0)
	signed, _ := r.Bool()
	signature, _ := r.String(0)
	strict, _ := r.Bool()
	if name != "Yun_mo_an" || count != 1 || key != "textures" || value != "texture-data" || !signed || signature != "signed-texture" || strict || r.Remaining() != 0 {
		t.Fatalf("wrong authenticated profile: name=%q count=%d property=%q/%q signature=%q strict=%t", name, count, key, value, signature, strict)
	}
	select {
	case got := <-query:
		if got != "Yun_mo_an:"+sessionHash(secret, publicDER) {
			t.Fatalf("hasJoined query = %q", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("hasJoined was not called")
	}
	client.Conn.Close()
	<-done
	if !s.Cfg.OnlineMode {
		t.Fatal("server was not in online mode")
	}
}

func TestOnlineLoginRejectsMissingSession(t *testing.T) {
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer httpServer.Close()
	_, client, done := onlineLoginPipe(t, httpServer.URL, "Tester")
	defer client.Conn.Close()
	answerEncryptionRequest(t, client, false)
	id, payload, err := client.ReadPacket()
	if err != nil || id != proto.SLoginDisconnect || !strings.Contains(string(payload), "Failed to verify username!") {
		t.Fatalf("rejected login: id=%d payload=%q err=%v", id, payload, err)
	}
	client.Conn.Close()
	<-done
}

func TestOnlineLoginRejectsWrongVerifyTokenBeforeSessionRequest(t *testing.T) {
	called := make(chan struct{}, 1)
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called <- struct{}{}
		w.WriteHeader(http.StatusOK)
	}))
	defer httpServer.Close()
	_, client, done := onlineLoginPipe(t, httpServer.URL, "Tester")
	defer client.Conn.Close()
	answerEncryptionRequest(t, client, true)
	id, _, err := client.ReadPacket()
	if err != nil || id != proto.SLoginDisconnect {
		t.Fatalf("wrong token response: id=%d err=%v", id, err)
	}
	select {
	case <-called:
		t.Fatal("session service contacted despite wrong verify token")
	default:
	}
	client.Conn.Close()
	<-done
}

func onlineLoginPipe(t *testing.T, endpoint, name string) (*Server, *proto.FramedConn, chan struct{}) {
	t.Helper()
	serverConn, clientConn := net.Pipe()
	_ = clientConn.SetDeadline(time.Now().Add(5 * time.Second))
	s := New(Config{MaxPlayers: 20, OnlineMode: true}, hub.New(), nil)
	s.auth.endpoint = endpoint
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer serverConn.Close()
		s.handleLogin(proto.NewFramedConn(serverConn), ProtocolVersion)
	}()
	client := proto.NewFramedConn(clientConn)
	if err := client.WritePacket(proto.CLoginStart, proto.NewWriter().String(name).UUID([16]byte{}).Payload()); err != nil {
		t.Fatal(err)
	}
	return s, client, done
}

func answerEncryptionRequest(t *testing.T, client *proto.FramedConn, wrongToken bool) ([]byte, []byte) {
	t.Helper()
	id, payload, err := client.ReadPacket()
	if err != nil || id != proto.SLoginEncryption {
		t.Fatalf("encryption request: id=%d err=%v", id, err)
	}
	r := proto.NewReader(payload)
	serverID, _ := r.String(0)
	keyLength, _ := r.VarInt()
	publicDER, err := r.Bytes(int(keyLength))
	if err != nil {
		t.Fatal(err)
	}
	tokenLength, _ := r.VarInt()
	token, err := r.Bytes(int(tokenLength))
	if err != nil {
		t.Fatal(err)
	}
	shouldAuthenticate, _ := r.Bool()
	if serverID != "" || keyLength < 128 || tokenLength != 4 || !shouldAuthenticate || r.Remaining() != 0 {
		t.Fatalf("malformed encryption request: serverID=%q key=%d token=%d authenticate=%t", serverID, keyLength, tokenLength, shouldAuthenticate)
	}
	parsed, err := x509.ParsePKIXPublicKey(publicDER)
	if err != nil {
		t.Fatal(err)
	}
	publicKey := parsed.(*rsa.PublicKey)
	secret := []byte("0123456789abcdef")
	secretCipher, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, secret)
	if err != nil {
		t.Fatal(err)
	}
	if wrongToken {
		token[0] ^= 0xff
	}
	tokenCipher, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, token)
	if err != nil {
		t.Fatal(err)
	}
	response := proto.NewWriter().VarInt(int32(len(secretCipher))).Raw(secretCipher).
		VarInt(int32(len(tokenCipher))).Raw(tokenCipher).Payload()
	if err := client.WritePacket(proto.CLoginEncryption, response); err != nil {
		t.Fatal(err)
	}
	if !wrongToken {
		if err := client.EnableEncryption(secret); err != nil {
			t.Fatal(err)
		}
	}
	return secret, publicDER
}
