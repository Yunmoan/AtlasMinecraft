import os, struct, json, sys

def rvarint(d, p):
    v=0
    for i in range(5):
        c=d[p]; p+=1
        v |= (c&0x7F)<<(7*i)
        if not c&0x80: return v,p
    raise ValueError

def rstr(d, p):
    n, p = rvarint(d, p)
    return d[p:p+n].decode('utf-8'), p+n

class P:
    def __init__(self, d, p=0): self.d=d; self.p=p
    def u8(self): v=self.d[self.p]; self.p+=1; return v
    def i32(self): v=struct.unpack('>i',self.d[self.p:self.p+4])[0]; self.p+=4; return v
    def s(self):
        n, p = rvarint(self.d, self.p)
        v = self.d[p:p+n].decode('utf-8'); self.p = p+n; return v
    def payload(self, t):
        d,p = self.d, self.p
        if t==1: v=struct.unpack('>b',d[p:p+1])[0]; self.p+=1; return v
        if t==2: v=struct.unpack('>h',d[p:p+2])[0]; self.p+=2; return v
        if t==3: v=struct.unpack('>i',d[p:p+4])[0]; self.p+=4; return v
        if t==4: v=struct.unpack('>q',d[p:p+8])[0]; self.p+=8; return v
        if t==5: v=struct.unpack('>f',d[p:p+4])[0]; self.p+=4; return v
        if t==6: v=struct.unpack('>d',d[p:p+8])[0]; self.p+=8; return v
        if t==7: n=self.i32(); v=d[self.p:self.p+n]; self.p+=n; return ('bytes',n)
        if t==8:
            n=struct.unpack('>H',d[p:p+2])[0]; self.p+=2+n; return d[p+2:p+2+n].decode('utf-8')
        if t==9:
            et=self.u8(); n=self.i32(); return [self.payload(et) for _ in range(n)]
        if t==10:
            out={}
            while True:
                tt=self.u8()
                if tt==0: return out
                ln=struct.unpack('>H',d[self.p:self.p+2])[0]; self.p+=2
                k=d[self.p:self.p+ln].decode('utf-8'); self.p+=ln
                out[k]=self.payload(tt)
        if t==11:
            n=self.i32(); self.p+=4*n; return ('iarr',n)
        if t==12:
            n=self.i32(); self.p+=8*n; return ('larr',n)
        raise ValueError('tag %d at %d' % (t,self.p))

def parse_registry_packet(d):
    reg, p = rstr(d, 0)
    n, p = rvarint(d, p)
    entries = {}
    for _ in range(n):
        name, p = rstr(d, p)
        present = d[p]; p += 1
        if present:
            pr = P(d, p)
            assert pr.u8() == 0x0A, 'root tag %x' % d[p]
            el = pr.payload(10)
            p = pr.p
            entries[name] = el
        else:
            entries[name] = None
    return reg, entries, p, len(d)

def load_dir(dirname):
    out = {}
    order = []
    for f in sorted(os.listdir(dirname)):
        if 'config-pkt07' not in f: continue
        d = open(os.path.join(dirname, f), 'rb').read()
        reg, entries, consumed, total = parse_registry_packet(d)
        out[reg] = entries
        order.append(reg)
        if consumed != total:
            print('  NOTE trailing bytes in %s: %d' % (f, total-consumed))
    return out, order

def typed(v):
    if isinstance(v, dict): return {k: typed(x) for k,x in v.items()}
    if isinstance(v, list): return [typed(x) for x in v]
    if isinstance(v, float): return ('f', round(v, 4))
    return v

def diff(a, b, path, out):
    if type(a) != type(b):
        out.append('%s: TYPE %s(%r) vs %s(%r)' % (path, type(a).__name__, a, type(b).__name__, b))
        return
    if isinstance(a, dict):
        for k in a:
            if k not in b: out.append('%s.%s: only in A' % (path, k)); continue
            diff(a[k], b[k], path+'.'+k, out)
        for k in b:
            if k not in a: out.append('%s.%s: only in B (%r)' % (path, k, b[k]))
    elif isinstance(a, list):
        if len(a) != len(b):
            out.append('%s: LEN %d vs %d' % (path, len(a), len(b)))
        for i,(x,y) in enumerate(zip(a,b)):
            diff(x, y, '%s[%d]' % (path,i), out)
    else:
        if typed(a) != typed(b):
            out.append('%s: %r vs %r' % (path, a, b))

van, van_order = load_dir('capture-vanilla')
print('vanilla registries:')
for r in van_order: print('  %-45s %d entries' % (r, len(van[r])))

mine_path = sys.argv[1] if len(sys.argv)>1 else None
if mine_path:
    mine, mine_order = load_dir(mine_path)
    print('mine registries:')
    for r in mine_order: print('  %-45s %d entries' % (r, len(mine[r])))
    print()
    print('=== only in vanilla ===')
    for r in van_order:
        if r not in mine: print(' REGISTRY', r)
    print('=== only in mine ===')
    for r in mine_order:
        if r not in van: print(' REGISTRY', r)
    print('=== content diffs (shared registries) ===')
    total = 0
    for r in van_order:
        if r not in mine: continue
        # diff first entry fully + entry key sets
        va, ma = van[r], mine[r]
        vkeys, mkeys = list(va.keys()), list(ma.keys())
        if vkeys != mkeys:
            print('%s: entry order/count differs: vanilla=%d mine=%d' % (r, len(vkeys), len(mkeys)))
        if vkeys and mkeys and vkeys[0] in ma:
            out = []
            diff(typed(va[vkeys[0]]), typed(ma[vkeys[0]]), r+'/'+vkeys[0], out)
            for line in out[:25]:
                print(' ', line)
            total += len(out)
    print('total first-entry diffs:', total)
