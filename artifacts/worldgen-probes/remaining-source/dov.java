/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.google.common.collect.ImmutableMap
 *  com.mojang.serialization.MapCodec
 *  javax.annotation.Nullable
 */
import com.google.common.collect.ImmutableMap;
import com.mojang.serialization.MapCodec;
import java.util.Map;
import java.util.function.Function;
import java.util.stream.Collectors;
import javax.annotation.Nullable;

public class dov
extends dfy {
    public static final MapCodec<dov> a = dov.b(dov::new);
    public static final dtt b = dlm.f;
    public static final dtt c = dlm.b;
    public static final dtt d = dlm.c;
    public static final dtt e = dlm.d;
    public static final dtt f = dlm.e;
    public static final Map<ji, dtt> g = dlm.h.entrySet().stream().filter($$0 -> $$0.getKey() != ji.a).collect(ad.a());
    protected static final float h = 1.0f;
    private static final exv i = dfy.a(0.0, 15.0, 0.0, 16.0, 16.0, 16.0);
    private static final exv j = dfy.a(0.0, 0.0, 0.0, 1.0, 16.0, 16.0);
    private static final exv k = dfy.a(15.0, 0.0, 0.0, 16.0, 16.0, 16.0);
    private static final exv l = dfy.a(0.0, 0.0, 0.0, 16.0, 16.0, 1.0);
    private static final exv m = dfy.a(0.0, 0.0, 15.0, 16.0, 16.0, 16.0);
    private final Map<dtc, exv> n;

    public MapCodec<dov> a() {
        return a;
    }

    public dov(dtb.d $$0) {
        super($$0);
        this.k((dtc)((dtc)((dtc)((dtc)((dtc)((dtc)this.E.b()).a(b, false)).a(c, false)).a(d, false)).a(e, false)).a(f, false));
        this.n = ImmutableMap.copyOf(this.E.a().stream().collect(Collectors.toMap(Function.identity(), dov::m)));
    }

    private static exv m(dtc $$0) {
        exv $$1 = exs.a();
        if ($$0.c(b).booleanValue()) {
            $$1 = i;
        }
        if ($$0.c(c).booleanValue()) {
            $$1 = exs.a($$1, l);
        }
        if ($$0.c(e).booleanValue()) {
            $$1 = exs.a($$1, m);
        }
        if ($$0.c(d).booleanValue()) {
            $$1 = exs.a($$1, k);
        }
        if ($$0.c(f).booleanValue()) {
            $$1 = exs.a($$1, j);
        }
        return $$1.c() ? exs.b() : $$1;
    }

    @Override
    protected exv a(dtc $$0, dcc $$1, jd $$2, exh $$3) {
        return this.n.get($$0);
    }

    @Override
    protected boolean a_(dtc $$0, dcc $$1, jd $$2) {
        return true;
    }

    @Override
    protected boolean a(dtc $$0, dcz $$1, jd $$2) {
        return this.n(this.i($$0, $$1, $$2));
    }

    private boolean n(dtc $$0) {
        return this.o($$0) > 0;
    }

    private int o(dtc $$0) {
        int $$1 = 0;
        for (dtt $$2 : g.values()) {
            if (!$$0.c($$2).booleanValue()) continue;
            ++$$1;
        }
        return $$1;
    }

    private boolean b(dcc $$0, jd $$1, ji $$2) {
        if ($$2 == ji.a) {
            return false;
        }
        jd $$3 = $$1.a($$2);
        if (dov.a($$0, $$3, $$2)) {
            return true;
        }
        if ($$2.o() != ji.a.b) {
            dtt $$4 = g.get($$2);
            dtc $$5 = $$0.a_($$1.d());
            return $$5.a(this) && $$5.c($$4) != false;
        }
        return false;
    }

    public static boolean a(dcc $$0, jd $$1, ji $$2) {
        return dky.a($$0, $$2, $$1, $$0.a_($$1));
    }

    private dtc i(dtc $$0, dcc $$1, jd $$2) {
        jd $$3 = $$2.d();
        if ($$0.c(b).booleanValue()) {
            $$0 = (dtc)$$0.a(b, dov.a($$1, $$3, ji.a));
        }
        dtb.a $$4 = null;
        for (ji $$5 : ji.c.a) {
            dtt $$6 = dov.a($$5);
            if (!$$0.c($$6).booleanValue()) continue;
            boolean $$7 = this.b($$1, $$2, $$5);
            if (!$$7) {
                if ($$4 == null) {
                    $$4 = $$1.a_($$3);
                }
                $$7 = $$4.a(this) && $$4.c($$6) != false;
            }
            $$0 = (dtc)$$0.a($$6, $$7);
        }
        return $$0;
    }

    @Override
    protected dtc a(dtc $$0, ji $$1, dtc $$2, dcx $$3, jd $$4, jd $$5) {
        if ($$1 == ji.a) {
            return super.a($$0, $$1, $$2, $$3, $$4, $$5);
        }
        dtc $$6 = this.i($$0, $$3, $$4);
        if (!this.n($$6)) {
            return dga.a.o();
        }
        return $$6;
    }

    @Override
    protected void b(dtc $$0, aqu $$1, jd $$2, ayw $$3) {
        dtc $$20;
        dtc $$19;
        jd $$17;
        dtc $$18;
        if (!$$1.ab().b(dcs.X)) {
            return;
        }
        if ($$3.a(4) != 0) {
            return;
        }
        ji $$4 = ji.b($$3);
        jd $$5 = $$2.d();
        if ($$4.o().d() && !$$0.c(dov.a($$4)).booleanValue()) {
            if (!this.a((dcc)$$1, $$2)) {
                return;
            }
            jd $$6 = $$2.a($$4);
            dtc $$7 = $$1.a_($$6);
            if ($$7.i()) {
                ji $$8 = $$4.h();
                ji $$9 = $$4.i();
                boolean $$10 = $$0.c(dov.a($$8));
                boolean $$11 = $$0.c(dov.a($$9));
                jd $$12 = $$6.a($$8);
                jd $$13 = $$6.a($$9);
                if ($$10 && dov.a((dcc)$$1, $$12, $$8)) {
                    $$1.a($$6, (dtc)this.o().a(dov.a($$8), true), 2);
                } else if ($$11 && dov.a((dcc)$$1, $$13, $$9)) {
                    $$1.a($$6, (dtc)this.o().a(dov.a($$9), true), 2);
                } else {
                    ji $$14 = $$4.g();
                    if ($$10 && $$1.u($$12) && dov.a((dcc)$$1, $$2.a($$8), $$14)) {
                        $$1.a($$12, (dtc)this.o().a(dov.a($$14), true), 2);
                    } else if ($$11 && $$1.u($$13) && dov.a((dcc)$$1, $$2.a($$9), $$14)) {
                        $$1.a($$13, (dtc)this.o().a(dov.a($$14), true), 2);
                    } else if ((double)$$3.i() < 0.05 && dov.a((dcc)$$1, $$6.d(), ji.b)) {
                        $$1.a($$6, (dtc)this.o().a(b, true), 2);
                    }
                }
            } else if (dov.a((dcc)$$1, $$6, $$4)) {
                $$1.a($$2, (dtc)$$0.a(dov.a($$4), true), 2);
            }
            return;
        }
        if ($$4 == ji.b && $$2.v() < $$1.am() - 1) {
            if (this.b($$1, $$2, $$4)) {
                $$1.a($$2, (dtc)$$0.a(b, true), 2);
                return;
            }
            if ($$1.u($$5)) {
                if (!this.a((dcc)$$1, $$2)) {
                    return;
                }
                dtc $$15 = $$0;
                for (ji $$16 : ji.c.a) {
                    if (!$$3.h() && dov.a((dcc)$$1, $$5.a($$16), $$16)) continue;
                    $$15 = (dtc)$$15.a(dov.a($$16), false);
                }
                if (this.p($$15)) {
                    $$1.a($$5, $$15, 2);
                }
                return;
            }
        }
        if ($$2.v() > $$1.I_() && (($$18 = $$1.a_($$17 = $$2.e())).i() || $$18.a(this)) && ($$19 = $$18.i() ? this.o() : $$18) != ($$20 = this.a($$0, $$19, $$3)) && this.p($$20)) {
            $$1.a($$17, $$20, 2);
        }
    }

    private dtc a(dtc $$0, dtc $$1, ayw $$2) {
        for (ji $$3 : ji.c.a) {
            dtt $$4;
            if (!$$2.h() || !$$0.c($$4 = dov.a($$3)).booleanValue()) continue;
            $$1 = (dtc)$$1.a($$4, true);
        }
        return $$1;
    }

    private boolean p(dtc $$0) {
        return $$0.c(c) != false || $$0.c(d) != false || $$0.c(e) != false || $$0.c(f) != false;
    }

    private boolean a(dcc $$0, jd $$1) {
        int $$2 = 4;
        Iterable<jd> $$3 = jd.b($$1.u() - 4, $$1.v() - 1, $$1.w() - 4, $$1.u() + 4, $$1.v() + 1, $$1.w() + 4);
        int $$4 = 5;
        for (jd $$5 : $$3) {
            if (!$$0.a_($$5).a(this) || --$$4 > 0) continue;
            return false;
        }
        return true;
    }

    @Override
    protected boolean a(dtc $$0, cyd $$1) {
        dtc $$2 = $$1.q().a_($$1.a());
        if ($$2.a(this)) {
            return this.o($$2) < g.size();
        }
        return super.a($$0, $$1);
    }

    @Override
    @Nullable
    public dtc a(cyd $$0) {
        dtc $$1 = $$0.q().a_($$0.a());
        boolean $$2 = $$1.a(this);
        dtc $$3 = $$2 ? $$1 : this.o();
        for (ji $$4 : $$0.f()) {
            boolean $$6;
            if ($$4 == ji.a) continue;
            dtt $$5 = dov.a($$4);
            boolean bl = $$6 = $$2 && $$1.c($$5) != false;
            if ($$6 || !this.b($$0.q(), $$0.a(), $$4)) continue;
            return (dtc)$$3.a($$5, true);
        }
        return $$2 ? $$3 : null;
    }

    @Override
    protected void a(dtd.a<dfy, dtc> $$0) {
        $$0.a(b, c, d, e, f);
    }

    @Override
    protected dtc a(dtc $$0, dmm $$1) {
        switch ($$1) {
            case c: {
                return (dtc)((dtc)((dtc)((dtc)$$0.a(c, $$0.c(e))).a(d, $$0.c(f))).a(e, $$0.c(c))).a(f, $$0.c(d));
            }
            case d: {
                return (dtc)((dtc)((dtc)((dtc)$$0.a(c, $$0.c(d))).a(d, $$0.c(e))).a(e, $$0.c(f))).a(f, $$0.c(c));
            }
            case b: {
                return (dtc)((dtc)((dtc)((dtc)$$0.a(c, $$0.c(f))).a(d, $$0.c(c))).a(e, $$0.c(d))).a(f, $$0.c(e));
            }
        }
        return $$0;
    }

    @Override
    protected dtc a(dtc $$0, dkv $$1) {
        switch ($$1) {
            case b: {
                return (dtc)((dtc)$$0.a(c, $$0.c(e))).a(e, $$0.c(c));
            }
            case c: {
                return (dtc)((dtc)$$0.a(d, $$0.c(f))).a(f, $$0.c(d));
            }
        }
        return super.a($$0, $$1);
    }

    public static dtt a(ji $$0) {
        return g.get($$0);
    }
}
