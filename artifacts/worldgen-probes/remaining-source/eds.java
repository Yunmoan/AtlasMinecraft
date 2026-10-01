/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  ayw
 *  com.mojang.serialization.Codec
 *  dcx
 *  ddc
 *  dds
 *  dga
 *  dtc
 *  dyo
 *  ece
 *  ecg
 *  efg
 *  ejj
 *  jd
 *  ji
 *  ji$c
 *  kh
 */
import com.mojang.serialization.Codec;
import java.util.Optional;
import java.util.OptionalInt;
import java.util.function.Predicate;

public class eds
extends ece<efg> {
    public eds(Codec<efg> $$0) {
        super($$0);
    }

    public boolean a(ecg<efg> $$0) {
        kh $$7;
        dds $$12 = $$0.b();
        jd $$22 = $$0.e();
        efg $$3 = (efg)$$0.f();
        ayw $$4 = $$0.d();
        OptionalInt $$5 = eds.a($$12, $$22, $$3);
        if ($$5.isEmpty()) {
            return false;
        }
        jd $$6 = $$22.h($$5.getAsInt());
        ejj $$8 = ejj.a((kh)$$6.b($$7 = new kh($$3.c, $$3.c, $$3.c)), (kh)$$6.a($$7));
        return jd.a((ejj)$$8).filter($$2 -> $$4.i() < $$1.d).filter($$1 -> this.b($$12, (jd)$$1)).mapToInt($$1 -> {
            $$12.a($$1, dga.kJ.o(), 2);
            return 1;
        }).sum() > 0;
    }

    private static OptionalInt a(dds $$02, jd $$1, efg $$2) {
        Predicate<dtc> $$3 = $$0 -> $$0.a(dga.G);
        Predicate<dtc> $$4 = $$0 -> !$$0.a(dga.G);
        Optional $$5 = dyo.a((ddc)$$02, (jd)$$1, (int)$$2.b, $$3, $$4);
        return $$5.map(dyo::c).orElseGet(OptionalInt::empty);
    }

    private boolean b(dds $$0, jd $$1) {
        if (this.a((dcx)$$0, $$1) || this.a((dcx)$$0, $$1.e())) {
            return false;
        }
        for (ji $$2 : ji.c.a) {
            if (!this.a((dcx)$$0, $$1.a($$2))) continue;
            return false;
        }
        return true;
    }

    private boolean a(dcx $$0, jd $$1) {
        dtc $$2 = $$0.a_($$1);
        return $$2.a(dga.G) || $$2.i();
    }
}
