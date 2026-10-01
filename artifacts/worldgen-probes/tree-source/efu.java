/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.datafixers.Products$P2
 *  com.mojang.datafixers.kinds.App
 *  com.mojang.serialization.Codec
 *  com.mojang.serialization.codecs.RecordCodecBuilder$Instance
 *  com.mojang.serialization.codecs.RecordCodecBuilder$Mu
 */
import com.mojang.datafixers.Products;
import com.mojang.datafixers.kinds.App;
import com.mojang.serialization.Codec;
import com.mojang.serialization.codecs.RecordCodecBuilder;

public abstract class efu {
    public static final Codec<efu> d = lt.U.r().dispatch(efu::a, efv::a);
    protected final bpw e;
    protected final bpw f;

    protected static <P extends efu> Products.P2<RecordCodecBuilder.Mu<P>, bpw, bpw> b(RecordCodecBuilder.Instance<P> $$02) {
        return $$02.group((App)bpw.b(0, 16).fieldOf("radius").forGetter($$0 -> $$0.e), (App)bpw.b(0, 16).fieldOf("offset").forGetter($$0 -> $$0.f));
    }

    public efu(bpw $$0, bpw $$1) {
        this.e = $$0;
        this.f = $$1;
    }

    protected abstract efv<?> a();

    public void a(ddc $$0, b $$1, ayw $$2, efe $$3, int $$4, a $$5, int $$6, int $$7) {
        this.a($$0, $$1, $$2, $$3, $$4, $$5, $$6, $$7, this.a($$2));
    }

    protected abstract void a(ddc var1, b var2, ayw var3, efe var4, int var5, a var6, int var7, int var8, int var9);

    public abstract int a(ayw var1, int var2, efe var3);

    public int a(ayw $$0, int $$1) {
        return this.e.a($$0);
    }

    private int a(ayw $$0) {
        return this.f.a($$0);
    }

    protected abstract boolean a(ayw var1, int var2, int var3, int var4, int var5, boolean var6);

    protected boolean b(ayw $$0, int $$1, int $$2, int $$3, int $$4, boolean $$5) {
        int $$9;
        int $$8;
        if ($$5) {
            int $$6 = Math.min(Math.abs($$1), Math.abs($$1 - 1));
            int $$7 = Math.min(Math.abs($$3), Math.abs($$3 - 1));
        } else {
            $$8 = Math.abs($$1);
            $$9 = Math.abs($$3);
        }
        return this.a($$0, $$8, $$2, $$9, $$4, $$5);
    }

    protected void a(ddc $$0, b $$1, ayw $$2, efe $$3, jd $$4, int $$5, int $$6, boolean $$7) {
        int $$8 = $$7 ? 1 : 0;
        jd.a $$9 = new jd.a();
        for (int $$10 = -$$5; $$10 <= $$5 + $$8; ++$$10) {
            for (int $$11 = -$$5; $$11 <= $$5 + $$8; ++$$11) {
                if (this.b($$2, $$10, $$6, $$11, $$5, $$7)) continue;
                $$9.a($$4, $$10, $$6, $$11);
                efu.a($$0, $$1, $$2, $$3, $$9);
            }
        }
    }

    protected final void a(ddc $$0, b $$1, ayw $$2, efe $$3, jd $$4, int $$5, int $$6, boolean $$7, float $$8, float $$9) {
        this.a($$0, $$1, $$2, $$3, $$4, $$5, $$6, $$7);
        int $$10 = $$7 ? 1 : 0;
        jd $$11 = $$4.e();
        jd.a $$12 = new jd.a();
        for (ji $$13 : ji.c.a) {
            ji $$14 = $$13.h();
            int $$15 = $$14.f() == ji.b.a ? $$5 + $$10 : $$5;
            $$12.a($$4, 0, $$6 - 1, 0).c($$14, $$15).c($$13, -$$5);
            for (int $$16 = -$$5; $$16 < $$5 + $$10; ++$$16) {
                boolean $$17 = $$1.a($$12.c(ji.b));
                $$12.c(ji.a);
                if ($$17 && efu.a($$0, $$1, $$2, $$3, $$8, $$11, $$12)) {
                    $$12.c(ji.a);
                    efu.a($$0, $$1, $$2, $$3, $$9, $$11, $$12);
                    $$12.c(ji.b);
                }
                $$12.c($$13);
            }
        }
    }

    private static boolean a(ddc $$0, b $$1, ayw $$2, efe $$3, float $$4, jd $$5, jd.a $$6) {
        if ($$6.k($$5) >= 7) {
            return false;
        }
        if ($$2.i() > $$4) {
            return false;
        }
        return efu.a($$0, $$1, $$2, $$3, $$6);
    }

    protected static boolean a(ddc $$02, b $$1, ayw $$2, efe $$3, jd $$4) {
        if (!edq.c($$02, $$4)) {
            return false;
        }
        dtc $$5 = $$3.e.a($$2, $$4);
        if ($$5.b(dts.C)) {
            $$5 = (dtc)$$5.a(dts.C, $$02.b($$4, $$0 -> $$0.a(epf.c)));
        }
        $$1.a($$4, $$5);
        return true;
    }

    public static interface b {
        public void a(jd var1, dtc var2);

        public boolean a(jd var1);
    }

    public static final class a {
        private final jd a;
        private final int b;
        private final boolean c;

        public a(jd $$0, int $$1, boolean $$2) {
            this.a = $$0;
            this.b = $$1;
            this.c = $$2;
        }

        public jd a() {
            return this.a;
        }

        public int b() {
            return this.b;
        }

        public boolean c() {
            return this.c;
        }
    }
}
