/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.serialization.MapCodec
 *  javax.annotation.Nullable
 */
import com.mojang.serialization.MapCodec;
import javax.annotation.Nullable;

public class dfc
extends dfy
implements dgb {
    public static final MapCodec<dfc> a = dfc.b(dfc::new);
    protected static final float b = 3.0f;
    protected static final float c = 5.0f;
    protected static final float d = 1.5f;
    protected static final exv e = dfy.a(5.0, 0.0, 5.0, 11.0, 16.0, 11.0);
    protected static final exv f = dfy.a(3.0, 0.0, 3.0, 13.0, 16.0, 13.0);
    protected static final exv g = dfy.a(6.5, 0.0, 6.5, 9.5, 16.0, 9.5);
    public static final duc h = dts.aq;
    public static final dua<dto> i = dts.bk;
    public static final duc j = dts.aU;
    public static final int k = 16;
    public static final int l = 0;
    public static final int m = 1;
    public static final int n = 0;
    public static final int o = 1;

    public MapCodec<dfc> a() {
        return a;
    }

    public dfc(dtb.d $$0) {
        super($$0);
        this.k((dtc)((dtc)((dtc)((dtc)this.E.b()).a(h, 0)).a(i, dto.a)).a(j, 0));
    }

    @Override
    protected void a(dtd.a<dfy, dtc> $$0) {
        $$0.a(h, i, j);
    }

    @Override
    protected boolean a_(dtc $$0, dcc $$1, jd $$2) {
        return true;
    }

    @Override
    protected exv a(dtc $$0, dcc $$1, jd $$2, exh $$3) {
        exv $$4 = $$0.c(i) == dto.c ? f : e;
        exc $$5 = $$0.n($$1, $$2);
        return $$4.a($$5.c, $$5.d, $$5.e);
    }

    @Override
    protected boolean a(dtc $$0, ept $$1) {
        return false;
    }

    @Override
    protected exv b(dtc $$0, dcc $$1, jd $$2, exh $$3) {
        exc $$4 = $$0.n($$1, $$2);
        return g.a($$4.c, $$4.d, $$4.e);
    }

    @Override
    protected boolean c(dtc $$0, dcc $$1, jd $$2) {
        return false;
    }

    @Override
    @Nullable
    public dtc a(cyd $$0) {
        epe $$1 = $$0.q().b_($$0.a());
        if (!$$1.c()) {
            return null;
        }
        dtc $$2 = $$0.q().a_($$0.a().e());
        if ($$2.a(awe.av)) {
            if ($$2.a(dga.mY)) {
                return (dtc)this.o().a(h, 0);
            }
            if ($$2.a(dga.mZ)) {
                int $$3 = $$2.c(h) > 0 ? 1 : 0;
                return (dtc)this.o().a(h, $$3);
            }
            dtc $$4 = $$0.q().a_($$0.a().d());
            if ($$4.a(dga.mZ)) {
                return (dtc)this.o().a(h, $$4.c(h));
            }
            return dga.mY.o();
        }
        return null;
    }

    @Override
    protected void a(dtc $$0, aqu $$1, jd $$2, ayw $$3) {
        if (!$$0.a((dcz)$$1, $$2)) {
            $$1.b($$2, true);
        }
    }

    @Override
    protected boolean d_(dtc $$0) {
        return $$0.c(j) == 0;
    }

    @Override
    protected void b(dtc $$0, aqu $$1, jd $$2, ayw $$3) {
        int $$4;
        if ($$0.c(j) != 0) {
            return;
        }
        if ($$3.a(3) == 0 && $$1.u($$2.d()) && $$1.b($$2.d(), 0) >= 9 && ($$4 = this.b($$1, $$2) + 1) < 16) {
            this.a($$0, (dcw)$$1, $$2, $$3, $$4);
        }
    }

    @Override
    protected boolean a(dtc $$0, dcz $$1, jd $$2) {
        return $$1.a_($$2.e()).a(awe.av);
    }

    @Override
    protected dtc a(dtc $$0, ji $$1, dtc $$2, dcx $$3, jd $$4, jd $$5) {
        if (!$$0.a($$3, $$4)) {
            $$3.a($$4, this, 1);
        }
        if ($$1 == ji.b && $$2.a(dga.mZ) && $$2.c(h) > $$0.c(h)) {
            $$3.a($$4, (dtc)$$0.a(h), 2);
        }
        return super.a($$0, $$1, $$2, $$3, $$4, $$5);
    }

    @Override
    public boolean b(dcz $$0, jd $$1, dtc $$2) {
        int $$4;
        int $$3 = this.a((dcc)$$0, $$1);
        return $$3 + ($$4 = this.b($$0, $$1)) + 1 < 16 && $$0.a_($$1.b($$3)).c(j) != 1;
    }

    @Override
    public boolean a(dcw $$0, ayw $$1, jd $$2, dtc $$3) {
        return true;
    }

    @Override
    public void a(aqu $$0, ayw $$1, jd $$2, dtc $$3) {
        int $$4 = this.a((dcc)$$0, $$2);
        int $$5 = this.b($$0, $$2);
        int $$6 = $$4 + $$5 + 1;
        int $$7 = 1 + $$1.a(2);
        for (int $$8 = 0; $$8 < $$7; ++$$8) {
            jd $$9 = $$2.b($$4);
            dtc $$10 = $$0.a_($$9);
            if ($$6 >= 16 || $$10.c(j) == 1 || !$$0.u($$9.d())) {
                return;
            }
            this.a($$10, (dcw)$$0, $$9, $$1, $$6);
            ++$$4;
            ++$$6;
        }
    }

    @Override
    protected float a(dtc $$0, cmx $$1, dcc $$2, jd $$3) {
        if ($$1.eT().g() instanceof cwg) {
            return 1.0f;
        }
        return super.a($$0, $$1, $$2, $$3);
    }

    protected void a(dtc $$0, dcw $$1, jd $$2, ayw $$3, int $$4) {
        dtc $$5 = $$1.a_($$2.e());
        jd $$6 = $$2.c(2);
        dtc $$7 = $$1.a_($$6);
        dto $$8 = dto.a;
        if ($$4 >= 1) {
            if (!$$5.a(dga.mZ) || $$5.c(i) == dto.a) {
                $$8 = dto.b;
            } else if ($$5.a(dga.mZ) && $$5.c(i) != dto.a) {
                $$8 = dto.c;
                if ($$7.a(dga.mZ)) {
                    $$1.a($$2.e(), (dtc)$$5.a(i, dto.b), 3);
                    $$1.a($$6, (dtc)$$7.a(i, dto.a), 3);
                }
            }
        }
        int $$9 = $$0.c(h) == 1 || $$7.a(dga.mZ) ? 1 : 0;
        int $$10 = $$4 >= 11 && $$3.i() < 0.25f || $$4 == 15 ? 1 : 0;
        $$1.a($$2.d(), (dtc)((dtc)((dtc)this.o().a(h, $$9)).a(i, $$8)).a(j, $$10), 3);
    }

    protected int a(dcc $$0, jd $$1) {
        int $$2;
        for ($$2 = 0; $$2 < 16 && $$0.a_($$1.b($$2 + 1)).a(dga.mZ); ++$$2) {
        }
        return $$2;
    }

    protected int b(dcc $$0, jd $$1) {
        int $$2;
        for ($$2 = 0; $$2 < 16 && $$0.a_($$1.c($$2 + 1)).a(dga.mZ); ++$$2) {
        }
        return $$2;
    }
}
