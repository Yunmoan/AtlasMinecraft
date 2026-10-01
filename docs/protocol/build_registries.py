#!/usr/bin/env python
"""Builds the Registry Data packet payloads for protocol 767 (1.21.1).

Since 1.20.5 the client expects ONE Registry Data packet per registry:
    Identifier(registry) + Prefixed Array of {Identifier entry, Optional NBT}
Entry order defines numeric ids (0, 1, 2, ...). Element NBT mirrors the
datapack JSON with the same primitive types; network NBT roots are anonymous
(0x0A, no name) since 1.20.2.

Reads registry dumps + elements under docs/protocol/mcmeta/ and emits
internal/registry/registries.bin — a container of packet payloads:
    VarInt(count) then per packet: VarInt(length) + payload bytes.
"""
import json
import os
import struct

HERE = os.path.dirname(os.path.abspath(__file__))
OUT = os.path.normpath(os.path.join(HERE, '..', '..', 'internal', 'registry', 'registries.bin'))

# JSON keys whose codec expects a specific NBT primitive type regardless of
# how the number looks in JSON.
FLOAT_KEYS = {'temperature', 'downfall', 'exhaustion', 'item_model_index',
              'coordinate_scale', 'ambient_light', 'level'}
DOUBLE_KEYS = {'offset'}
LONG_KEYS = {'fixed_time'}


def varint(n):
    out = bytearray()
    while True:
        if n & ~0x7F == 0:
            out.append(n)
            return bytes(out)
        out.append((n & 0x7F) | 0x80)
        n >>= 7


def vstr(s):
    b = s.encode('utf-8')
    return varint(len(b)) + b


def infer_list_type(items):
    if not items:
        return 0  # TAG_End for empty lists — accepted by the vanilla reader
    types = {json_nbt_type(i) for i in items}
    if types == {3}:
        return 3
    if types <= {3, 5}:
        return 5
    if types == {8}:
        return 8
    if types == {10}:
        return 10
    if types == {9}:
        return 9
    raise ValueError('unsupported mixed list types: %s' % types)


def json_nbt_type(v):
    if isinstance(v, bool):
        return 1
    if isinstance(v, int):
        return 3
    if isinstance(v, float):
        return 5
    if isinstance(v, str):
        return 8
    if isinstance(v, list):
        return 9
    if isinstance(v, dict):
        return 10
    raise ValueError('unsupported json value: %r' % (v,))


class NBT:
    """Builds an anonymous-root network NBT compound."""

    def __init__(self):
        self.buf = bytearray(b'\x0a')  # TAG_Compound, no name since 1.20.2

    def finish(self):
        return bytes(self.buf + b'\x00')

    def tag(self, name, v):
        t = json_nbt_type(v)
        if isinstance(v, (int, float)) and not isinstance(v, bool):
            if name in FLOAT_KEYS:
                t = 5
            elif name in DOUBLE_KEYS:
                t = 6
            elif name in LONG_KEYS:
                t = 4
        self.buf += bytes([t]) + struct.pack('>H', len(name)) + name.encode()
        if t == 9:
            et = infer_list_type(v)
            self.buf += bytes([et]) + struct.pack('>i', len(v))
            for item in v:
                self.any(item, name)
        elif t == 10:
            for k, item in v.items():
                self.tag(k, item)
            self.buf += b'\x00'
        elif t == 1:
            self.buf += struct.pack('>b', 1 if v else 0)
        elif t == 3:
            self.buf += struct.pack('>i', v)
        elif t == 4:
            self.buf += struct.pack('>q', v)
        elif t == 5:
            self.buf += struct.pack('>f', v)
        elif t == 6:
            self.buf += struct.pack('>d', v)
        elif t == 8:
            b = v.encode('utf-8')
            self.buf += struct.pack('>H', len(b)) + b
        else:
            raise ValueError('tag %d' % t)
        return self

    def any(self, v, hint=None):
        t = json_nbt_type(v)
        if isinstance(v, bool):
            self.buf += struct.pack('>b', 1 if v else 0)
        elif isinstance(v, int):
            self.buf += struct.pack('>q' if hint in LONG_KEYS else '>i', v)
        elif isinstance(v, float):
            self.buf += struct.pack('>d' if hint in DOUBLE_KEYS else '>f', v)
        elif isinstance(v, str):
            b = v.encode()
            self.buf += struct.pack('>H', len(b)) + b
        elif isinstance(v, list):
            et = infer_list_type(v)
            self.buf += bytes([et]) + struct.pack('>i', len(v))
            for item in v:
                self.any(item, hint)
        elif isinstance(v, dict):
            for k, item in v.items():
                self.tag(k, item)
            self.buf += b'\x00'
        else:
            raise ValueError(repr(v))
        return self


def element_nbt(elem):
    n = NBT()
    for k, v in elem.items():
        n.tag(k, v)
    return n.finish()


def registry_packet(tag, entries):
    out = vstr(tag) + varint(len(entries))
    for key, elem in entries:
        out += vstr(key) + b'\x01' + element_nbt(elem)  # present + NBT
    return out


def main():
    mc = os.path.join(HERE, 'mcmeta')

    def load(p):
        with open(os.path.join(mc, p), encoding='utf-8') as f:
            return json.load(f)

    def elements(kind, prefix):
        out = []
        for name in load(prefix + '.json'):
            with open(os.path.join(mc, 'elements', kind, name + '.json'), encoding='utf-8') as f:
                out.append(('minecraft:' + name, json.load(f)))
        return out

    packets = [
        registry_packet('minecraft:dimension_type', elements('dimension_type', 'dimension_type')),
        registry_packet('minecraft:worldgen/biome', elements('biome', 'worldgen_biome')),
        registry_packet('minecraft:damage_type', elements('damage_type', 'damage_type')),
        registry_packet('minecraft:trim_pattern', elements('trim_pattern', 'trim_pattern')),
        registry_packet('minecraft:trim_material', elements('trim_material', 'trim_material')),
        registry_packet('minecraft:wolf_variant', elements('wolf_variant', 'wolf_variant')),
        registry_packet('minecraft:painting_variant', elements('painting_variant', 'painting_variant')),
        registry_packet('minecraft:banner_pattern', elements('banner_pattern', 'banner_pattern')),
    ]

    blob = varint(len(packets))
    for p in packets:
        blob += varint(len(p)) + p

    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, 'wb') as f:
        f.write(blob)
    print('wrote %s (%d bytes, %d packets)' % (OUT, len(blob), len(packets)))


if __name__ == '__main__':
    main()
