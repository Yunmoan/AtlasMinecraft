/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  ayw
 *  com.mojang.serialization.Codec
 *  dcx
 *  dds
 *  ece
 *  ecg
 *  eer
 *  jd
 *  ji
 *  ji$c
 */
import com.mojang.serialization.Codec;
import java.util.Optional;

public class eda
extends ece<eer> {
    public eda(Codec<eer> $$0) {
        super($$0);
    }

    public boolean a(ecg<eer> $$0) {
        dds $$1 = $$0.b();
        jd $$2 = $$0.e();
        ayw $$3 = $$0.d();
        eer $$4 = (eer)$$0.f();
        Optional<ji> $$5 = eda.a((dcx)$$1, $$2, $$3);
        if ($$5.isEmpty()) {
            return false;
        }
        jd $$6 = $$2.a($$5.get().g());
        eda.a((dcx)$$1, $$3, $$6, $$4);
        int $$7 = $$3.i() < $$4.b && ebz.c($$1.a_($$2.a($$5.get()))) ? 2 : 1;
        ebz.a((dcx)$$1, $$2, $$5.get(), $$7, false);
        return true;
    }

    private static Optional<ji> a(dcx $$0, jd $$1, ayw $$2) {
        boolean $$3 = ebz.b($$0.a_($$1.d()));
        boolean $$4 = ebz.b($$0.a_($$1.e()));
        if ($$3 && $$4) {
            return Optional.of($$2.h() ? ji.a : ji.b);
        }
        if ($$3) {
            return Optional.of(ji.a);
        }
        if ($$4) {
            return Optional.of(ji.b);
        }
        return Optional.empty();
    }

    private static void a(dcx $$0, ayw $$1, jd $$2, eer $$3) {
        ebz.c($$0, $$2);
        for (ji $$4 : ji.c.a) {
            if ($$1.i() > $$3.c) continue;
            jd $$5 = $$2.a($$4);
            ebz.c($$0, $$5);
            if ($$1.i() > $$3.d) continue;
            jd $$6 = $$5.a(ji.b((ayw)$$1));
            ebz.c($$0, $$6);
            if ($$1.i() > $$3.e) continue;
            jd $$7 = $$6.a(ji.b((ayw)$$1));
            ebz.c($$0, $$7);
        }
    }
}
