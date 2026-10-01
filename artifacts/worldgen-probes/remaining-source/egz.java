/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  ayw
 *  com.mojang.serialization.Codec
 *  com.mojang.serialization.MapCodec
 *  dov
 *  dtt
 *  eha
 *  eha$a
 *  ehb
 *  jd
 */
import com.mojang.serialization.Codec;
import com.mojang.serialization.MapCodec;

public class egz
extends eha {
    public static final MapCodec<egz> a = Codec.floatRange((float)0.0f, (float)1.0f).fieldOf("probability").xmap(egz::new, $$0 -> Float.valueOf($$0.b));
    private final float b;

    protected ehb<?> a() {
        return ehb.b;
    }

    public egz(float $$0) {
        this.b = $$0;
    }

    public void a(eha.a $$0) {
        ayw $$1 = $$0.b();
        $$0.d().forEach($$2 -> {
            jd $$6;
            jd $$5;
            jd $$4;
            jd $$3;
            if ($$1.i() < this.b && $$0.a($$3 = $$2.h())) {
                egz.a($$3, dov.d, $$0);
            }
            if ($$1.i() < this.b && $$0.a($$4 = $$2.i())) {
                egz.a($$4, dov.f, $$0);
            }
            if ($$1.i() < this.b && $$0.a($$5 = $$2.f())) {
                egz.a($$5, dov.e, $$0);
            }
            if ($$1.i() < this.b && $$0.a($$6 = $$2.g())) {
                egz.a($$6, dov.c, $$0);
            }
        });
    }

    private static void a(jd $$0, dtt $$1, eha.a $$2) {
        $$2.a($$0, $$1);
        $$0 = $$0.e();
        for (int $$3 = 4; $$2.a($$0) && $$3 > 0; --$$3) {
            $$2.a($$0, $$1);
            $$0 = $$0.e();
        }
    }
}
