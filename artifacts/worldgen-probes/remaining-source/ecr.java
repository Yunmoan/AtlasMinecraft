/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.serialization.Codec
 */
import com.mojang.serialization.Codec;

public class ecr
extends ece<eeb> {
    public ecr(Codec<eeb> $$0) {
        super($$0);
    }

    @Override
    public boolean a(ecg<eeb> $$0) {
        boolean $$24;
        int $$11;
        jd $$1 = $$0.e();
        dds $$2 = $$0.b();
        $$1 = new jd($$1.u(), $$0.c().f(), $$1.w());
        ayw $$3 = $$0.d();
        boolean $$4 = $$3.j() > 0.7;
        dtc $$5 = $$0.f().b;
        double $$6 = $$3.j() * 2.0 * Math.PI;
        int $$7 = 11 - $$3.a(5);
        int $$8 = 3 + $$3.a(3);
        boolean $$9 = $$3.j() > 0.7;
        int $$10 = 11;
        int n = $$11 = $$9 ? $$3.a(6) + 6 : $$3.a(15) + 3;
        if (!$$9 && $$3.j() > 0.9) {
            $$11 += $$3.a(19) + 7;
        }
        int $$12 = Math.min($$11 + $$3.a(11), 18);
        int $$13 = Math.min($$11 + $$3.a(7) - $$3.a(5), 11);
        int $$14 = $$9 ? $$7 : 11;
        for (int $$15 = -$$14; $$15 < $$14; ++$$15) {
            for (int $$16 = -$$14; $$16 < $$14; ++$$16) {
                for (int $$17 = 0; $$17 < $$11; ++$$17) {
                    int $$18;
                    int n2 = $$18 = $$9 ? this.b($$17, $$11, $$13) : this.a($$3, $$17, $$11, $$13);
                    if (!$$9 && $$15 >= $$18) continue;
                    this.a($$2, $$3, $$1, $$11, $$15, $$17, $$16, $$18, $$14, $$9, $$8, $$6, $$4, $$5);
                }
            }
        }
        this.a($$2, $$1, $$13, $$11, $$9, $$7);
        for (int $$19 = -$$14; $$19 < $$14; ++$$19) {
            for (int $$20 = -$$14; $$20 < $$14; ++$$20) {
                for (int $$21 = -1; $$21 > -$$12; --$$21) {
                    int $$22 = $$9 ? ayo.f((float)$$14 * (1.0f - (float)Math.pow($$21, 2.0) / ((float)$$12 * 8.0f))) : $$14;
                    int $$23 = this.b($$3, -$$21, $$12, $$13);
                    if ($$19 >= $$23) continue;
                    this.a($$2, $$3, $$1, $$12, $$19, $$21, $$20, $$23, $$22, $$9, $$8, $$6, $$4, $$5);
                }
            }
        }
        boolean bl = $$9 ? $$3.j() > 0.1 : ($$24 = $$3.j() > 0.7);
        if ($$24) {
            this.a($$3, $$2, $$13, $$11, $$1, $$9, $$7, $$6, $$8);
        }
        return true;
    }

    private void a(ayw $$0, dcx $$1, int $$2, int $$3, jd $$4, boolean $$5, int $$6, double $$7, int $$8) {
        int $$9 = $$0.h() ? -1 : 1;
        int $$10 = $$0.h() ? -1 : 1;
        int $$11 = $$0.a(Math.max($$2 / 2 - 2, 1));
        if ($$0.h()) {
            $$11 = $$2 / 2 + 1 - $$0.a(Math.max($$2 - $$2 / 2 - 1, 1));
        }
        int $$12 = $$0.a(Math.max($$2 / 2 - 2, 1));
        if ($$0.h()) {
            $$12 = $$2 / 2 + 1 - $$0.a(Math.max($$2 - $$2 / 2 - 1, 1));
        }
        if ($$5) {
            $$11 = $$12 = $$0.a(Math.max($$6 - 5, 1));
        }
        jd $$13 = new jd($$9 * $$11, 0, $$10 * $$12);
        double $$14 = $$5 ? $$7 + 1.5707963267948966 : $$0.j() * 2.0 * Math.PI;
        for (int $$15 = 0; $$15 < $$3 - 3; ++$$15) {
            int $$16 = this.a($$0, $$15, $$3, $$2);
            this.a($$16, $$15, $$4, $$1, false, $$14, $$13, $$6, $$8);
        }
        for (int $$17 = -1; $$17 > -$$3 + $$0.a(5); --$$17) {
            int $$18 = this.b($$0, -$$17, $$3, $$2);
            this.a($$18, $$17, $$4, $$1, true, $$14, $$13, $$6, $$8);
        }
    }

    private void a(int $$0, int $$1, jd $$2, dcx $$3, boolean $$4, double $$5, jd $$6, int $$7, int $$8) {
        int $$9 = $$0 + 1 + $$7 / 3;
        int $$10 = Math.min($$0 - 3, 3) + $$8 / 2 - 1;
        for (int $$11 = -$$9; $$11 < $$9; ++$$11) {
            for (int $$12 = -$$9; $$12 < $$9; ++$$12) {
                jd $$14;
                dtc $$15;
                double $$13 = this.a($$11, $$12, $$6, $$9, $$10, $$5);
                if (!($$13 < 0.0) || !ecr.c($$15 = $$3.a_($$14 = $$2.b($$11, $$1, $$12))) && !$$15.a(dga.dP)) continue;
                if ($$4) {
                    this.a($$3, $$14, dga.G.o());
                    continue;
                }
                this.a($$3, $$14, dga.a.o());
                this.a($$3, $$14);
            }
        }
    }

    private void a(dcx $$0, jd $$1) {
        if ($$0.a_($$1.d()).a(dga.dN)) {
            this.a($$0, $$1.d(), dga.a.o());
        }
    }

    private void a(dcx $$0, ayw $$1, jd $$2, int $$3, int $$4, int $$5, int $$6, int $$7, int $$8, boolean $$9, int $$10, double $$11, boolean $$12, dtc $$13) {
        double $$14;
        double d2 = $$14 = $$9 ? this.a($$4, $$6, jd.c, $$8, this.a($$5, $$3, $$10), $$11) : this.a($$4, $$6, jd.c, $$7, $$1);
        if ($$14 < 0.0) {
            double $$16;
            jd $$15 = $$2.b($$4, $$5, $$6);
            double d3 = $$16 = $$9 ? -0.5 : (double)(-6 - $$1.a(3));
            if ($$14 > $$16 && $$1.j() > 0.9) {
                return;
            }
            this.a($$15, $$0, $$1, $$3 - $$5, $$3, $$9, $$12, $$13);
        }
    }

    private void a(jd $$0, dcx $$1, ayw $$2, int $$3, int $$4, boolean $$5, boolean $$6, dtc $$7) {
        dtc $$8 = $$1.a_($$0);
        if ($$8.i() || $$8.a(dga.dP) || $$8.a(dga.dO) || $$8.a(dga.G)) {
            int $$10;
            boolean $$9 = !$$5 || $$2.j() > 0.05;
            int n = $$10 = $$5 ? 3 : 2;
            if ($$6 && !$$8.a(dga.G) && (double)$$3 <= (double)$$2.a(Math.max(1, $$4 / $$10)) + (double)$$4 * 0.6 && $$9) {
                this.a($$1, $$0, dga.dP.o());
            } else {
                this.a($$1, $$0, $$7);
            }
        }
    }

    private int a(int $$0, int $$1, int $$2) {
        int $$3 = $$2;
        if ($$0 > 0 && $$1 - $$0 <= 3) {
            $$3 -= 4 - ($$1 - $$0);
        }
        return $$3;
    }

    private double a(int $$0, int $$1, jd $$2, int $$3, ayw $$4) {
        float $$5 = 10.0f * ayo.a($$4.i(), 0.2f, 0.8f) / (float)$$3;
        return (double)$$5 + Math.pow($$0 - $$2.u(), 2.0) + Math.pow($$1 - $$2.w(), 2.0) - Math.pow($$3, 2.0);
    }

    private double a(int $$0, int $$1, jd $$2, int $$3, int $$4, double $$5) {
        return Math.pow(((double)($$0 - $$2.u()) * Math.cos($$5) - (double)($$1 - $$2.w()) * Math.sin($$5)) / (double)$$3, 2.0) + Math.pow(((double)($$0 - $$2.u()) * Math.sin($$5) + (double)($$1 - $$2.w()) * Math.cos($$5)) / (double)$$4, 2.0) - 1.0;
    }

    private int a(ayw $$0, int $$1, int $$2, int $$3) {
        float $$4 = 3.5f - $$0.i();
        float $$5 = (1.0f - (float)Math.pow($$1, 2.0) / ((float)$$2 * $$4)) * (float)$$3;
        if ($$2 > 15 + $$0.a(5)) {
            int $$6 = $$1 < 3 + $$0.a(6) ? $$1 / 2 : $$1;
            $$5 = (1.0f - (float)$$6 / ((float)$$2 * $$4 * 0.4f)) * (float)$$3;
        }
        return ayo.f($$5 / 2.0f);
    }

    private int b(int $$0, int $$1, int $$2) {
        float $$3 = 1.0f;
        float $$4 = (1.0f - (float)Math.pow($$0, 2.0) / ((float)$$1 * 1.0f)) * (float)$$2;
        return ayo.f($$4 / 2.0f);
    }

    private int b(ayw $$0, int $$1, int $$2, int $$3) {
        float $$4 = 1.0f + $$0.i() / 2.0f;
        float $$5 = (1.0f - (float)$$1 / ((float)$$2 * $$4)) * (float)$$3;
        return ayo.f($$5 / 2.0f);
    }

    private static boolean c(dtc $$0) {
        return $$0.a(dga.iC) || $$0.a(dga.dP) || $$0.a(dga.mW);
    }

    private boolean a(dcc $$0, jd $$1) {
        return $$0.a_($$1.e()).i();
    }

    private void a(dcx $$0, jd $$1, int $$2, int $$3, boolean $$4, int $$5) {
        int $$6 = $$4 ? $$5 : $$2 / 2;
        for (int $$7 = -$$6; $$7 <= $$6; ++$$7) {
            for (int $$8 = -$$6; $$8 <= $$6; ++$$8) {
                for (int $$9 = 0; $$9 <= $$3; ++$$9) {
                    jd $$10 = $$1.b($$7, $$9, $$8);
                    dtc $$11 = $$0.a_($$10);
                    if (!ecr.c($$11) && !$$11.a(dga.dN)) continue;
                    if (this.a((dcc)$$0, $$10)) {
                        this.a($$0, $$10, dga.a.o());
                        this.a($$0, $$10.d(), dga.a.o());
                        continue;
                    }
                    if (!ecr.c($$11)) continue;
                    dtc[] $$12 = new dtc[]{$$0.a_($$10.h()), $$0.a_($$10.i()), $$0.a_($$10.f()), $$0.a_($$10.g())};
                    int $$13 = 0;
                    for (dtc $$14 : $$12) {
                        if (ecr.c($$14)) continue;
                        ++$$13;
                    }
                    if ($$13 < 3) continue;
                    this.a($$0, $$10, dga.a.o());
                }
            }
        }
    }
}
