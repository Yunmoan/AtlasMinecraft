/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  ab
 *  ad
 *  akq
 *  arb
 *  ayo
 *  ayw
 *  com.google.common.annotations.VisibleForTesting
 *  com.google.common.base.Suppliers
 *  com.google.common.collect.Sets
 *  com.mojang.datafixers.kinds.App
 *  com.mojang.datafixers.kinds.Applicative
 *  com.mojang.serialization.MapCodec
 *  com.mojang.serialization.codecs.RecordCodecBuilder
 *  dcd
 *  dcy
 *  ddh
 *  ddi
 *  ddl
 *  ddq
 *  ddw
 *  ddx
 *  ddy
 *  ddz
 *  dea
 *  dga
 *  dtc
 *  dux
 *  duy
 *  duz
 *  dvj
 *  dvs
 *  dwy
 *  dyk
 *  dyk$a
 *  dyk$b
 *  dyl
 *  dym
 *  dyr$b
 *  dyr$e
 *  dys$b
 *  dys$c
 *  dyu$a
 *  dyy
 *  dyy$a
 *  dyz
 *  dzc
 *  dzd
 *  dze
 *  dzf
 *  dzg
 *  dzm
 *  dzn
 *  dzv
 *  dzx
 *  eaa
 *  eaz
 *  ebc
 *  javax.annotation.Nullable
 *  jd
 *  jd$a
 *  jm
 *  jx
 *  jz
 *  lu
 *  org.apache.commons.lang3.mutable.MutableObject
 */
import com.google.common.annotations.VisibleForTesting;
import com.google.common.base.Suppliers;
import com.google.common.collect.Sets;
import com.mojang.datafixers.kinds.App;
import com.mojang.datafixers.kinds.Applicative;
import com.mojang.serialization.MapCodec;
import com.mojang.serialization.codecs.RecordCodecBuilder;
import java.text.DecimalFormat;
import java.util.HashSet;
import java.util.List;
import java.util.OptionalInt;
import java.util.concurrent.CompletableFuture;
import java.util.function.Predicate;
import java.util.function.Supplier;
import javax.annotation.Nullable;
import org.apache.commons.lang3.mutable.MutableObject;

public final class dzb
extends duz {
    public static final MapCodec<dzb> c = RecordCodecBuilder.mapCodec($$02 -> $$02.group((App)dea.a.fieldOf("biome_source").forGetter($$0 -> $$0.b), (App)dzd.b.fieldOf("settings").forGetter($$0 -> $$0.e)).apply((Applicative)$$02, $$02.stable(dzb::new)));
    private static final dtc d = dga.a.o();
    private final jm<dzd> e;
    private final Supplier<dyk.a> f;

    public dzb(dea $$0, jm<dzd> $$1) {
        super($$0);
        this.e = $$1;
        this.f = Suppliers.memoize(() -> dzb.a((dzd)$$1.a()));
    }

    private static dyk.a a(dzd $$0) {
        dyk.b $$1 = new dyk.b(-54, dga.H.o());
        int $$2 = $$0.l();
        dyk.b $$3 = new dyk.b($$2, $$0.h());
        dyk.b $$42 = new dyk.b(dwy.e * 2, dga.a.o());
        return ($$4, $$5, $$6) -> {
            if ($$5 < Math.min(-54, $$2)) {
                return $$1;
            }
            return $$3;
        };
    }

    public CompletableFuture<duy> a(dzm $$0, eaa $$1, ddq $$2, duy $$3) {
        return CompletableFuture.supplyAsync(ad.a((String)"init_biomes", () -> {
            this.b($$1, $$0, $$2, $$3);
            return $$3;
        }), ad.g());
    }

    private void b(eaa $$0, dzm $$1, ddq $$2, duy $$32) {
        dzc $$4 = $$32.a((T $$3) -> this.a((duy)$$3, $$2, $$0, $$1));
        ddz $$5 = dym.a((ddz)$$0.a((ddz)this.b), (duy)$$32);
        $$32.a($$5, $$4.a($$1.a(), ((dzd)this.e.a()).k()));
    }

    private dzc a(duy $$0, ddq $$1, eaa $$2, dzm $$3) {
        return dzc.a((duy)$$0, (dzm)$$3, (dys.c)dyl.a((ddq)$$1, (dcd)$$0.f()), (dzd)((dzd)this.e.a()), (dyk.a)this.f.get(), (eaa)$$2);
    }

    protected MapCodec<? extends duz> b() {
        return c;
    }

    public jm<dzd> h() {
        return this.e;
    }

    public boolean a(akq<dzd> $$0) {
        return this.e.a($$0);
    }

    public int a(int $$0, int $$1, dyy.a $$2, dcy $$3, dzm $$4) {
        return this.a($$3, $$4, $$0, $$1, null, $$2.e()).orElse($$3.I_());
    }

    public ddi a(int $$0, int $$1, dcy $$2, dzm $$3) {
        MutableObject $$4 = new MutableObject();
        this.a($$2, $$3, $$0, $$1, (MutableObject<ddi>)$$4, null);
        return (ddi)$$4.getValue();
    }

    public void a(List<String> $$0, dzm $$1, jd $$2) {
        DecimalFormat $$3 = new DecimalFormat("0.000");
        dze $$4 = $$1.a();
        dyr.e $$5 = new dyr.e($$2.u(), $$2.v(), $$2.w());
        double $$6 = $$4.j().a((dyr.b)$$5);
        $$0.add("NoiseRouter T: " + $$3.format($$4.e().a((dyr.b)$$5)) + " V: " + $$3.format($$4.f().a((dyr.b)$$5)) + " C: " + $$3.format($$4.g().a((dyr.b)$$5)) + " E: " + $$3.format($$4.h().a((dyr.b)$$5)) + " D: " + $$3.format($$4.i().a((dyr.b)$$5)) + " W: " + $$3.format($$6) + " PV: " + $$3.format(dzf.a((float)((float)$$6))) + " AS: " + $$3.format($$4.k().a((dyr.b)$$5)) + " N: " + $$3.format($$4.l().a((dyr.b)$$5)));
    }

    private OptionalInt a(dcy $$0, dzm $$1, int $$2, int $$3, @Nullable MutableObject<ddi> $$4, @Nullable Predicate<dtc> $$5) {
        dtc[] $$12;
        dzg $$6 = ((dzd)this.e.a()).f().a($$0);
        int $$7 = $$6.a();
        int $$8 = $$6.c();
        int $$9 = ayo.a((int)$$8, (int)$$7);
        int $$10 = ayo.a((int)$$6.d(), (int)$$7);
        if ($$10 <= 0) {
            return OptionalInt.empty();
        }
        if ($$4 == null) {
            Object $$11 = null;
        } else {
            $$12 = new dtc[$$6.d()];
            $$4.setValue((Object)new ddi($$8, $$12));
        }
        int $$13 = $$6.b();
        int $$14 = Math.floorDiv($$2, $$13);
        int $$15 = Math.floorDiv($$3, $$13);
        int $$16 = Math.floorMod($$2, $$13);
        int $$17 = Math.floorMod($$3, $$13);
        int $$18 = $$14 * $$13;
        int $$19 = $$15 * $$13;
        double $$20 = (double)$$16 / (double)$$13;
        double $$21 = (double)$$17 / (double)$$13;
        dzc $$22 = new dzc(1, $$1, $$18, $$19, $$6, (dys.c)dys.b.a, (dzd)this.e.a(), this.f.get(), eaa.a());
        $$22.f();
        $$22.b(0);
        for (int $$23 = $$10 - 1; $$23 >= 0; --$$23) {
            $$22.b($$23, 0);
            for (int $$24 = $$7 - 1; $$24 >= 0; --$$24) {
                dtc $$28;
                int $$25 = ($$9 + $$23) * $$7 + $$24;
                double $$26 = (double)$$24 / (double)$$7;
                $$22.a($$25, $$26);
                $$22.b($$2, $$20);
                $$22.c($$3, $$21);
                dtc $$27 = $$22.e();
                dtc dtc2 = $$28 = $$27 == null ? ((dzd)this.e.a()).g() : $$27;
                if ($$12 != null) {
                    int $$29 = $$23 * $$7 + $$24;
                    $$12[$$29] = $$28;
                }
                if ($$5 == null || !$$5.test($$28)) continue;
                $$22.g();
                return OptionalInt.of($$25 + 1);
            }
        }
        $$22.g();
        return OptionalInt.empty();
    }

    public void a(arb $$0, ddq $$1, dzm $$2, duy $$3) {
        if (ab.a((dcd)$$3.f())) {
            return;
        }
        dzv $$4 = new dzv((duz)this, (dcy)$$0);
        this.a($$3, $$4, $$2, $$1, $$0.F_(), (jz<ddw>)$$0.H_().d(lu.aF), eaa.a((arb)$$0));
    }

    @VisibleForTesting
    public void a(duy $$0, dzv $$1, dzm $$2, ddq $$32, ddy $$4, jz<ddw> $$5, eaa $$6) {
        dzc $$7 = $$0.a((T $$3) -> this.a((duy)$$3, $$32, $$6, $$2));
        dzd $$8 = (dzd)this.e.a();
        $$2.c().a($$2, $$4, $$5, $$8.n(), $$1, $$0, $$7, $$8.j());
    }

    public void a(arb $$0, long $$12, dzm $$22, ddy $$32, ddq $$4, duy $$5, dyu.a $$6) {
        ddy $$7 = $$32.a(($$1, $$2, $$3) -> this.b.getNoiseBiome($$1, $$2, $$3, $$22.b()));
        dzx $$8 = new dzx((ayw)new dyz(dzn.a()));
        int $$9 = 8;
        dcd $$10 = $$5.f();
        dzc $$11 = $$5.a((T $$3) -> this.a((duy)$$3, $$4, eaa.a((arb)$$0), $$22));
        dyk $$122 = $$11.i();
        eaz $$13 = new eaz(this, $$0.H_(), $$5.z(), $$11, $$22, ((dzd)this.e.a()).j());
        dux $$14 = ((dvs)$$5).b($$6);
        for (int $$15 = -8; $$15 <= 8; ++$$15) {
            for (int $$16 = -8; $$16 <= 8; ++$$16) {
                dcd $$17 = new dcd($$10.e + $$15, $$10.f + $$16);
                duy $$18 = $$0.a($$17.e, $$17.f);
                ddx $$19 = $$18.a(() -> this.a(this.b.getNoiseBiome(jx.a((int)$$17.d()), 0, jx.a((int)$$17.e()), $$22.b())));
                Iterable $$20 = $$19.a($$6);
                int $$21 = 0;
                for (jm $$222 : $$20) {
                    ebc $$23 = (ebc)$$222.a();
                    $$8.c($$12 + (long)$$21, $$17.e, $$17.f);
                    if ($$23.a((ayw)$$8)) {
                        $$23.a($$13, $$5, arg_0 -> ((ddy)$$7).a(arg_0), (ayw)$$8, $$122, $$17, $$14);
                    }
                    ++$$21;
                }
            }
        }
    }

    public CompletableFuture<duy> a(eaa $$0, dzm $$1, ddq $$2, duy $$3) {
        dzg $$4 = ((dzd)this.e.a()).f().a($$3.z());
        int $$5 = $$4.c();
        int $$6 = ayo.a((int)$$5, (int)$$4.a());
        int $$7 = ayo.a((int)$$4.d(), (int)$$4.a());
        if ($$7 <= 0) {
            return CompletableFuture.completedFuture($$3);
        }
        return CompletableFuture.supplyAsync(ad.a((String)"wgen_fill_noise", () -> {
            int $$8 = $$3.e($$7 * $$4.a() - 1 + $$5);
            int $$9 = $$3.e($$5);
            HashSet $$10 = Sets.newHashSet();
            for (int $$11 = $$8; $$11 >= $$9; --$$11) {
                dvj $$12 = $$3.b($$11);
                $$12.a();
                $$10.add($$12);
            }
            try {
                duy duy2 = this.a($$0, $$2, $$1, $$3, $$6, $$7);
                return duy2;
            }
            finally {
                for (dvj $$13 : $$10) {
                    $$13.b();
                }
            }
        }), ad.g());
    }

    private duy a(eaa $$0, ddq $$1, dzm $$2, duy $$32, int $$4, int $$5) {
        dzc $$6 = $$32.a((T $$3) -> this.a((duy)$$3, $$1, $$0, $$2));
        dyy $$7 = $$32.a(dyy.a.c);
        dyy $$8 = $$32.a(dyy.a.a);
        dcd $$9 = $$32.f();
        int $$10 = $$9.d();
        int $$11 = $$9.e();
        dyk $$12 = $$6.i();
        $$6.f();
        jd.a $$13 = new jd.a();
        int $$14 = $$6.j();
        int $$15 = $$6.k();
        int $$16 = 16 / $$14;
        int $$17 = 16 / $$14;
        for (int $$18 = 0; $$18 < $$16; ++$$18) {
            $$6.b($$18);
            for (int $$19 = 0; $$19 < $$17; ++$$19) {
                int $$20 = $$32.an() - 1;
                dvj $$21 = $$32.b($$20);
                for (int $$22 = $$5 - 1; $$22 >= 0; --$$22) {
                    $$6.b($$22, $$19);
                    for (int $$23 = $$15 - 1; $$23 >= 0; --$$23) {
                        int $$24 = ($$4 + $$22) * $$15 + $$23;
                        int $$25 = $$24 & 0xF;
                        int $$26 = $$32.e($$24);
                        if ($$20 != $$26) {
                            $$20 = $$26;
                            $$21 = $$32.b($$26);
                        }
                        double $$27 = (double)$$23 / (double)$$15;
                        $$6.a($$24, $$27);
                        for (int $$28 = 0; $$28 < $$14; ++$$28) {
                            int $$29 = $$10 + $$18 * $$14 + $$28;
                            int $$30 = $$29 & 0xF;
                            double $$31 = (double)$$28 / (double)$$14;
                            $$6.b($$29, $$31);
                            for (int $$322 = 0; $$322 < $$14; ++$$322) {
                                int $$33 = $$11 + $$19 * $$14 + $$322;
                                int $$34 = $$33 & 0xF;
                                double $$35 = (double)$$322 / (double)$$14;
                                $$6.c($$33, $$35);
                                dtc $$36 = $$6.e();
                                if ($$36 == null) {
                                    $$36 = ((dzd)this.e.a()).g();
                                }
                                if (($$36 = this.a($$6, $$29, $$24, $$33, $$36)) == d || ab.a((dcd)$$32.f())) continue;
                                $$21.a($$30, $$25, $$34, $$36, false);
                                $$7.a($$30, $$24, $$34, $$36);
                                $$8.a($$30, $$24, $$34, $$36);
                                if (!$$12.a() || $$36.u().c()) continue;
                                $$13.d($$29, $$24, $$33);
                                $$32.e((jd)$$13);
                            }
                        }
                    }
                }
            }
            $$6.h();
        }
        $$6.g();
        return $$32;
    }

    private dtc a(dzc $$0, int $$1, int $$2, int $$3, dtc $$4) {
        return $$4;
    }

    public int e() {
        return ((dzd)this.e.a()).f().d();
    }

    public int f() {
        return ((dzd)this.e.a()).l();
    }

    public int g() {
        return ((dzd)this.e.a()).f().c();
    }

    public void a(arb $$0) {
        if (((dzd)this.e.a()).a()) {
            return;
        }
        dcd $$1 = $$0.a();
        jm $$2 = $$0.t($$1.l().h($$0.am() - 1));
        dzx $$3 = new dzx((ayw)new dyz(dzn.a()));
        $$3.a($$0.C(), $$1.d(), $$1.e());
        ddh.a((ddl)$$0, (jm)$$2, (dcd)$$1, (ayw)$$3);
    }
}
