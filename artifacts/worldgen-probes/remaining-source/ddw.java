/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  ad
 *  akn
 *  akq
 *  avm
 *  avo
 *  ayo
 *  ayw
 *  com.google.common.collect.ImmutableList
 *  com.mojang.datafixers.kinds.App
 *  com.mojang.datafixers.kinds.Applicative
 *  com.mojang.serialization.Codec
 *  com.mojang.serialization.codecs.RecordCodecBuilder
 *  dcq
 *  dcu
 *  dcz
 *  ddf
 *  ddt
 *  ddu
 *  ddv
 *  ddw$1
 *  ddw$b
 *  ddw$c
 *  ddx
 *  dec
 *  dei
 *  dga
 *  dko
 *  dtc
 *  dyz
 *  dzx
 *  eoj
 *  epe
 *  epf
 *  it.unimi.dsi.fastutil.longs.Long2FloatLinkedOpenHashMap
 *  jd
 *  jm
 *  jq
 *  kb
 *  lu
 */
import com.google.common.collect.ImmutableList;
import com.mojang.datafixers.kinds.App;
import com.mojang.datafixers.kinds.Applicative;
import com.mojang.serialization.Codec;
import com.mojang.serialization.codecs.RecordCodecBuilder;
import it.unimi.dsi.fastutil.longs.Long2FloatLinkedOpenHashMap;
import java.util.List;
import java.util.Optional;

public final class ddw {
    public static final Codec<ddw> a = RecordCodecBuilder.create($$02 -> $$02.group((App)ddw$b.a.forGetter($$0 -> $$0.i), (App)dec.a.fieldOf("effects").forGetter($$0 -> $$0.l), (App)ddx.b.forGetter($$0 -> $$0.j), (App)dei.c.forGetter($$0 -> $$0.k)).apply((Applicative)$$02, ddw::new));
    public static final Codec<ddw> b = RecordCodecBuilder.create($$02 -> $$02.group((App)ddw$b.a.forGetter($$0 -> $$0.i), (App)dec.a.fieldOf("effects").forGetter($$0 -> $$0.l)).apply((Applicative)$$02, ($$0, $$1) -> new ddw((b)$$0, (dec)$$1, ddx.a, dei.b)));
    public static final Codec<jm<ddw>> c = akn.a((akq)lu.aF, a);
    public static final Codec<jq<ddw>> d = kb.a((akq)lu.aF, a);
    private static final eoj f = new eoj((ayw)new dzx((ayw)new dyz(1234L)), (List)ImmutableList.of((Object)0));
    static final eoj g = new eoj((ayw)new dzx((ayw)new dyz(3456L)), (List)ImmutableList.of((Object)-2, (Object)-1, (Object)0));
    @Deprecated(forRemoval=true)
    public static final eoj e = new eoj((ayw)new dzx((ayw)new dyz(2345L)), (List)ImmutableList.of((Object)0));
    private static final int h = 1024;
    private final b i;
    private final ddx j;
    private final dei k;
    private final dec l;
    private final ThreadLocal<Long2FloatLinkedOpenHashMap> m = ThreadLocal.withInitial(() -> (Long2FloatLinkedOpenHashMap)ad.a(() -> {
        1 $$0 = new /* Unavailable Anonymous Inner Class!! */;
        $$0.defaultReturnValue(Float.NaN);
        return $$0;
    }));

    ddw(b $$0, dec $$1, ddx $$2, dei $$3) {
        this.i = $$0;
        this.j = $$2;
        this.k = $$3;
        this.l = $$1;
    }

    public int a() {
        return this.l.d();
    }

    public dei b() {
        return this.k;
    }

    public boolean c() {
        return this.i.a();
    }

    public c a(jd $$0) {
        if (!this.c()) {
            return ddw$c.a;
        }
        return this.b($$0) ? ddw$c.c : ddw$c.b;
    }

    private float e(jd $$0) {
        float $$1 = this.i.d.a($$0, this.g());
        if ($$0.v() > 80) {
            float $$2 = (float)(f.a((double)((float)$$0.u() / 8.0f), (double)((float)$$0.w() / 8.0f), false) * 8.0);
            return $$1 - ($$2 + (float)$$0.v() - 80.0f) * 0.05f / 40.0f;
        }
        return $$1;
    }

    @Deprecated
    private float f(jd $$0) {
        long $$1 = $$0.a();
        Long2FloatLinkedOpenHashMap $$2 = this.m.get();
        float $$3 = $$2.get($$1);
        if (!Float.isNaN($$3)) {
            return $$3;
        }
        float $$4 = this.e($$0);
        if ($$2.size() == 1024) {
            $$2.removeFirstFloat();
        }
        $$2.put($$1, $$4);
        return $$4;
    }

    public boolean a(dcz $$0, jd $$1) {
        return this.a($$0, $$1, true);
    }

    public boolean a(dcz $$0, jd $$1, boolean $$2) {
        if (this.c($$1)) {
            return false;
        }
        if ($$1.v() >= $$0.I_() && $$1.v() < $$0.am() && $$0.a(ddf.b, $$1) < 10) {
            dtc $$3 = $$0.a_($$1);
            epe $$4 = $$0.b_($$1);
            if ($$4.a() == epf.c && $$3.b() instanceof dko) {
                boolean $$5;
                if (!$$2) {
                    return true;
                }
                boolean bl = $$5 = $$0.z($$1.h()) && $$0.z($$1.i()) && $$0.z($$1.f()) && $$0.z($$1.g());
                if (!$$5) {
                    return true;
                }
            }
        }
        return false;
    }

    public boolean b(jd $$0) {
        return !this.c($$0);
    }

    public boolean c(jd $$0) {
        return this.f($$0) >= 0.15f;
    }

    public boolean d(jd $$0) {
        return this.f($$0) > 0.1f;
    }

    public boolean b(dcz $$0, jd $$1) {
        dtc $$2;
        if (this.c($$1)) {
            return false;
        }
        return $$1.v() >= $$0.I_() && $$1.v() < $$0.am() && $$0.a(ddf.b, $$1) < 10 && (($$2 = $$0.a_($$1)).i() || $$2.a(dga.dN)) && dga.dN.o().a($$0, $$1);
    }

    public ddx d() {
        return this.j;
    }

    public int e() {
        return this.l.a();
    }

    public int a(double $$0, double $$1) {
        int $$2 = this.l.f().orElseGet(this::p);
        return this.l.g().a($$0, $$1, $$2);
    }

    private int p() {
        double $$0 = ayo.a((float)this.i.c, (float)0.0f, (float)1.0f);
        double $$1 = ayo.a((float)this.i.e, (float)0.0f, (float)1.0f);
        return dcu.a((double)$$0, (double)$$1);
    }

    public int f() {
        return this.l.e().orElseGet(this::q);
    }

    private int q() {
        double $$0 = ayo.a((float)this.i.c, (float)0.0f, (float)1.0f);
        double $$1 = ayo.a((float)this.i.e, (float)0.0f, (float)1.0f);
        return dcq.a((double)$$0, (double)$$1);
    }

    public float g() {
        return this.i.c;
    }

    public dec h() {
        return this.l;
    }

    public int i() {
        return this.l.b();
    }

    public int j() {
        return this.l.c();
    }

    public Optional<ddv> k() {
        return this.l.h();
    }

    public Optional<jm<avo>> l() {
        return this.l.i();
    }

    public Optional<ddu> m() {
        return this.l.j();
    }

    public Optional<ddt> n() {
        return this.l.k();
    }

    public Optional<avm> o() {
        return this.l.l();
    }
}
