/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.serialization.Codec
 */
import com.mojang.serialization.Codec;
import java.util.HashSet;
import java.util.Set;
import java.util.function.Predicate;

public class edw
extends edt {
    public edw(Codec<efh> $$0) {
        super($$0);
    }

    @Override
    protected Set<jd> a(dds $$0, efh $$1, ayw $$2, jd $$3, Predicate<dtc> $$4, int $$5, int $$6) {
        Set<jd> $$7 = super.a($$0, $$1, $$2, $$3, $$4, $$5, $$6);
        HashSet<jd> $$8 = new HashSet<jd>();
        jd.a $$9 = new jd.a();
        for (jd $$10 : $$7) {
            if (edw.a($$0, $$7, $$10, $$9)) continue;
            $$8.add($$10);
        }
        for (jd $$11 : $$8) {
            $$0.a($$11, dga.G.o(), 2);
        }
        return $$8;
    }

    private static boolean a(dds $$0, Set<jd> $$1, jd $$2, jd.a $$3) {
        return edw.a($$0, $$2, $$3, ji.c) || edw.a($$0, $$2, $$3, ji.f) || edw.a($$0, $$2, $$3, ji.d) || edw.a($$0, $$2, $$3, ji.e) || edw.a($$0, $$2, $$3, ji.a);
    }

    private static boolean a(dds $$0, jd $$1, jd.a $$2, ji $$3) {
        $$2.a((kh)$$1, $$3);
        return !$$0.a_($$2).d($$0, $$2, $$3.g());
    }

    @Override
    protected boolean a(dds $$0, efh $$1, duz $$2, ayw $$3, jd $$4) {
        if (super.a($$0, $$1, $$2, $$3, $$4.e())) {
            dtc $$5 = $$0.a_($$4);
            if ($$5.b(dts.C) && !$$5.c(dts.C).booleanValue()) {
                $$0.a($$4, (dtc)$$5.a(dts.C, true), 2);
            }
            return true;
        }
        return false;
    }
}
