/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  ayo
 *  ayw
 *  bpw
 *  com.mojang.datafixers.kinds.App
 *  com.mojang.datafixers.kinds.Applicative
 *  com.mojang.serialization.MapCodec
 *  com.mojang.serialization.codecs.RecordCodecBuilder
 *  com.mojang.serialization.codecs.RecordCodecBuilder$Instance
 *  ddc
 *  efe
 *  efu
 *  efu$a
 *  efu$b
 *  efv
 *  jd
 */
import com.mojang.datafixers.kinds.App;
import com.mojang.datafixers.kinds.Applicative;
import com.mojang.serialization.MapCodec;
import com.mojang.serialization.codecs.RecordCodecBuilder;

public class efx
extends efu {
    public static final MapCodec<efx> a = RecordCodecBuilder.mapCodec($$02 -> efx.b((RecordCodecBuilder.Instance)$$02).and((App)bpw.b((int)0, (int)24).fieldOf("crown_height").forGetter($$0 -> $$0.b)).apply((Applicative)$$02, efx::new));
    private final bpw b;

    public efx(bpw $$0, bpw $$1, bpw $$2) {
        super($$0, $$1);
        this.b = $$2;
    }

    protected efv<?> a() {
        return efv.h;
    }

    protected void a(ddc $$0, efu.b $$1, ayw $$2, efe $$3, int $$4, efu.a $$5, int $$6, int $$7, int $$8) {
        jd $$9 = $$5.a();
        int $$10 = 0;
        for (int $$11 = $$9.v() - $$6 + $$8; $$11 <= $$9.v() + $$8; ++$$11) {
            int $$15;
            int $$12 = $$9.v() - $$11;
            int $$13 = $$7 + $$5.b() + ayo.d((float)((float)$$12 / (float)$$6 * 3.5f));
            if ($$12 > 0 && $$13 == $$10 && ($$11 & 1) == 0) {
                int $$14 = $$13 + 1;
            } else {
                $$15 = $$13;
            }
            this.a($$0, $$1, $$2, $$3, new jd($$9.u(), $$11, $$9.w()), $$15, 0, $$5.c());
            $$10 = $$13;
        }
    }

    public int a(ayw $$0, int $$1, efe $$2) {
        return this.b.a($$0);
    }

    protected boolean a(ayw $$0, int $$1, int $$2, int $$3, int $$4, boolean $$5) {
        if ($$1 + $$3 >= 7) {
            return true;
        }
        return $$1 * $$1 + $$3 * $$3 > $$4 * $$4;
    }
}
