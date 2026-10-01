/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.serialization.Codec
 */
import com.mojang.serialization.Codec;

public class edc
extends ece<eev> {
    public edc(Codec<eev> $$0) {
        super($$0);
    }

    @Override
    public boolean a(ecg<eev> $$0) {
        eev $$1 = $$0.f();
        ayw $$2 = $$0.d();
        jd $$3 = $$0.e();
        dds $$4 = $$0.b();
        int $$5 = 0;
        jd.a $$6 = new jd.a();
        int $$7 = $$1.b() + 1;
        int $$8 = $$1.c() + 1;
        for (int $$9 = 0; $$9 < $$1.a(); ++$$9) {
            $$6.a($$3, $$2.a($$7) - $$2.a($$7), $$2.a($$8) - $$2.a($$8), $$2.a($$7) - $$2.a($$7));
            if (!$$1.d().a().a($$4, $$0.c(), $$2, $$6)) continue;
            ++$$5;
        }
        return $$5 > 0;
    }
}

