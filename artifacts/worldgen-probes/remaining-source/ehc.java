/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  ayw
 *  com.mojang.serialization.MapCodec
 *  dov
 *  eha
 *  eha$a
 *  ehb
 *  jd
 */
import com.mojang.serialization.MapCodec;

public class ehc
extends eha {
    public static final MapCodec<ehc> a = MapCodec.unit(() -> b);
    public static final ehc b = new ehc();

    protected ehb<?> a() {
        return ehb.a;
    }

    public void a(eha.a $$0) {
        ayw $$1 = $$0.b();
        $$0.c().forEach($$2 -> {
            jd $$6;
            jd $$5;
            jd $$4;
            jd $$3;
            if ($$1.a(3) > 0 && $$0.a($$3 = $$2.h())) {
                $$0.a($$3, dov.d);
            }
            if ($$1.a(3) > 0 && $$0.a($$4 = $$2.i())) {
                $$0.a($$4, dov.f);
            }
            if ($$1.a(3) > 0 && $$0.a($$5 = $$2.f())) {
                $$0.a($$5, dov.e);
            }
            if ($$1.a(3) > 0 && $$0.a($$6 = $$2.g())) {
                $$0.a($$6, dov.c);
            }
        });
    }
}
