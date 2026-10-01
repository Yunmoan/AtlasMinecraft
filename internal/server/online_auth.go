package server

import (
	"atlas/internal/player"
	"atlas/internal/proto"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/subtle"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const mojangHasJoinedURL = "https://sessionserver.mojang.com/session/minecraft/hasJoined"

var errInvalidSession = errors.New("failed to verify username")

type onlineAuthentication struct {
	once      sync.Once
	key       *rsa.PrivateKey
	publicDER []byte
	keyErr    error
	client    *http.Client
	endpoint  string
}

type onlineProfile struct {
	Name       string
	UUID       [16]byte
	Properties []player.ProfileProperty
}

func (a *onlineAuthentication) ensureKey() error {
	a.once.Do(func() {
		a.key, a.keyErr = rsa.GenerateKey(rand.Reader, 1024) // vanilla's RSA key size
		if a.keyErr == nil {
			a.publicDER, a.keyErr = x509.MarshalPKIXPublicKey(&a.key.PublicKey)
		}
	})
	return a.keyErr
}

// authenticate performs the 1.21.1 Encryption Request / Response exchange.
// The response is read in cleartext; every subsequent packet uses AES/CFB8.
func (a *onlineAuthentication) authenticate(fc *proto.FramedConn, name string) (onlineProfile, error) {
	if err := a.ensureKey(); err != nil {
		return onlineProfile{}, err
	}
	var token [4]byte
	if _, err := rand.Read(token[:]); err != nil {
		return onlineProfile{}, err
	}
	request := proto.NewWriter().String("").VarInt(int32(len(a.publicDER))).Raw(a.publicDER).
		VarInt(int32(len(token))).Raw(token[:]).Bool(true).Payload()
	if err := fc.WritePacket(proto.SLoginEncryption, request); err != nil {
		return onlineProfile{}, err
	}
	_ = fc.Conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	id, payload, err := fc.ReadPacket()
	if err != nil {
		return onlineProfile{}, err
	}
	if id != proto.CLoginEncryption {
		return onlineProfile{}, errors.New("expected Encryption Response")
	}
	r := proto.NewReader(payload)
	secretCipher, err := encryptedLoginField(r, a.key.Size())
	if err != nil {
		return onlineProfile{}, err
	}
	tokenCipher, err := encryptedLoginField(r, a.key.Size())
	if err != nil || r.Remaining() != 0 {
		return onlineProfile{}, errors.New("malformed Encryption Response")
	}
	secret, err := rsa.DecryptPKCS1v15(rand.Reader, a.key, secretCipher)
	if err != nil || len(secret) != 16 {
		return onlineProfile{}, errors.New("invalid encrypted shared secret")
	}
	answer, err := rsa.DecryptPKCS1v15(rand.Reader, a.key, tokenCipher)
	if err != nil || subtle.ConstantTimeCompare(answer, token[:]) != 1 {
		return onlineProfile{}, errors.New("invalid encryption verify token")
	}
	if err := fc.EnableEncryption(secret); err != nil {
		return onlineProfile{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return a.verifyJoined(ctx, name, sessionHash(secret, a.publicDER))
}

func encryptedLoginField(r *proto.Reader, size int) ([]byte, error) {
	n, err := r.VarInt()
	if err != nil || int(n) != size {
		return nil, errors.New("invalid encrypted field length")
	}
	return r.Bytes(int(n))
}

// Java's BigInteger(byte[]).toString(16) treats the SHA-1 digest as signed.
func sessionHash(secret, publicDER []byte) string {
	input := make([]byte, 0, len(secret)+len(publicDER))
	input = append(input, secret...)
	input = append(input, publicDER...)
	digest := sha1.Sum(input) // server ID is the empty string
	n := new(big.Int).SetBytes(digest[:])
	if digest[0]&0x80 != 0 {
		n.Sub(n, new(big.Int).Lsh(big.NewInt(1), 160))
	}
	return n.Text(16)
}

func (a *onlineAuthentication) verifyJoined(ctx context.Context, name, hash string) (onlineProfile, error) {
	endpoint := a.endpoint
	if endpoint == "" {
		endpoint = mojangHasJoinedURL
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return onlineProfile{}, err
	}
	q := u.Query()
	q.Set("username", name)
	q.Set("serverId", hash)
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return onlineProfile{}, err
	}
	req.Header.Set("Accept", "application/json")
	client := a.client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return onlineProfile{}, fmt.Errorf("session server request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound {
		return onlineProfile{}, errInvalidSession
	}
	if resp.StatusCode != http.StatusOK {
		return onlineProfile{}, fmt.Errorf("session server returned HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20+1))
	if err != nil {
		return onlineProfile{}, err
	}
	if len(body) > 1<<20 {
		return onlineProfile{}, errors.New("session profile is too large")
	}
	var joined struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Properties []struct {
			Name      string `json:"name"`
			Value     string `json:"value"`
			Signature string `json:"signature"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(body, &joined); err != nil {
		return onlineProfile{}, fmt.Errorf("parse session profile: %w", err)
	}
	if !validOperatorName(joined.Name) || !strings.EqualFold(joined.Name, name) {
		return onlineProfile{}, errInvalidSession
	}
	id, err := hex.DecodeString(strings.ReplaceAll(joined.ID, "-", ""))
	if err != nil || len(id) != 16 {
		return onlineProfile{}, errInvalidSession
	}
	profile := onlineProfile{Name: joined.Name}
	copy(profile.UUID[:], id)
	for _, property := range joined.Properties {
		if property.Name == "" || len(property.Name) > 32767 || len(property.Value) > 32767 || len(property.Signature) > 32767 {
			return onlineProfile{}, errInvalidSession
		}
		profile.Properties = append(profile.Properties, player.ProfileProperty{
			Name: property.Name, Value: property.Value, Signature: property.Signature,
		})
	}
	return profile, nil
}
