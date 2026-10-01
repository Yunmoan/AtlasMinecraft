/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.datafixers.kinds.App
 *  com.mojang.datafixers.kinds.Applicative
 *  com.mojang.serialization.MapCodec
 *  com.mojang.serialization.codecs.RecordCodecBuilder
 */
import com.mojang.datafixers.kinds.App;
import com.mojang.datafixers.kinds.Applicative;
import com.mojang.serialization.MapCodec;
import com.mojang.serialization.codecs.RecordCodecBuilder;
import java.util.Optional;

public class dla
extends dgh
implements dgb {
    public static final MapCodec<dla> a = RecordCodecBuilder.mapCodec($$02 -> $$02.group((App)akq.a(lu.aI).fieldOf("feature").forGetter($$0 -> $$0.d), dla.u()).apply((Applicative)$$02, dla::new));
    protected static final float b = 3.0f;
    protected static final exv c = dfy.a(5.0, 0.0, 5.0, 11.0, 6.0, 11.0);
    private final akq<ebq<?, ?>> d;

    public MapCodec<dla> a() {
        return a;
    }

    public dla(akq<ebq<?, ?>> $$0, dtb.d $$1) {
        super($$1);
        this.d = $$0;
    }

    @Override
    protected exv a(dtc $$0, dcc $$1, jd $$2, exh $$3) {
        return c;
    }

    @Override
    protected void b(dtc $$0, aqu $$1, jd $$2, ayw $$3) {
        if ($$3.a(25) == 0) {
            int $$4 = 5;
            int $$5 = 4;
            for (jd $$6 : jd.c($$2.b(-4, -1, -4), $$2.b(4, 1, 4))) {
                if (!$$1.a_($$6).a(this) || --$$4 > 0) continue;
                return;
            }
            jd $$7 = $$2.b($$3.a(3) - 1, $$3.a(2) - $$3.a(2), $$3.a(3) - 1);
            for (int $$8 = 0; $$8 < 4; ++$$8) {
                if ($$1.u($$7) && $$0.a((dcz)$$1, $$7)) {
                    $$2 = $$7;
                }
                $$7 = $$2.b($$3.a(3) - 1, $$3.a(2) - $$3.a(2), $$3.a(3) - 1);
            }
            if ($$1.u($$7) && $$0.a((dcz)$$1, $$7)) {
                $$1.a($$7, $$0, 2);
            }
        }
    }

    @Override
    protected boolean b(dtc $$0, dcc $$1, jd $$2) {
        return $$0.i($$1, $$2);
    }

    @Override
    protected boolean a(dtc $$0, dcz $$1, jd $$2) {
        jd $$3 = $$2.e();
        dtc $$4 = $$1.a_($$3);
        if ($$4.a(awe.bb)) {
            return true;
        }
        return $$1.b($$2, 0) < 13 && this.b($$4, $$1, $$3);
    }

    public boolean a(aqu $$0, jd $$1, dtc $$2, ayw $$3) {
        Optional<jm.c<ebq<?, ?>>> $$4 = $$0.H_().d(lu.aI).b(this.d);
        if ($$4.isEmpty()) {
            return false;
        }
        $$0.a($$1, false);
        if (((ebq)((jm)$$4.get()).a()).a($$0, $$0.l().g(), $$3, $$1)) {
            return true;
        }
        $$0.a($$1, $$2, 3);
        return false;
    }

    @Override
    public boolean b(dcz $$0, jd $$1, dtc $$2) {
        return true;
    }

    @Override
    public boolean a(dcw $$0, ayw $$1, jd $$2, dtc $$3) {
        return (double)$$1.i() < 0.4;
    }

    @Override
    public void a(aqu $$0, ayw $$1, jd $$2, dtc $$3) {
        this.a($$0, $$2, $$3, $$1);
    }
}
