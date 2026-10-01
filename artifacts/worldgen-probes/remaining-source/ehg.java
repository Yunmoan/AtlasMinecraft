/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  ayw
 *  com.google.common.collect.Lists
 *  com.mojang.datafixers.kinds.Applicative
 *  com.mojang.serialization.MapCodec
 *  com.mojang.serialization.codecs.RecordCodecBuilder
 *  com.mojang.serialization.codecs.RecordCodecBuilder$Instance
 *  ddc
 *  dtc
 *  edq
 *  efe
 *  efu$a
 *  ehm
 *  ehn
 *  jd
 *  ji
 *  ji$c
 */
import com.google.common.collect.Lists;
import com.mojang.datafixers.kinds.Applicative;
import com.mojang.serialization.MapCodec;
import com.mojang.serialization.codecs.RecordCodecBuilder;
import java.util.ArrayList;
import java.util.List;
import java.util.function.BiConsumer;

public class ehg
extends ehm {
    public static final MapCodec<ehg> a = RecordCodecBuilder.mapCodec($$0 -> ehg.a((RecordCodecBuilder.Instance)$$0).apply((Applicative)$$0, ehg::new));

    public ehg(int $$0, int $$1, int $$2) {
        super($$0, $$1, $$2);
    }

    protected ehn<?> a() {
        return ehn.e;
    }

    public List<efu.a> a(ddc $$0, BiConsumer<jd, dtc> $$1, ayw $$2, int $$3, jd $$4, efe $$5) {
        ArrayList $$6 = Lists.newArrayList();
        jd $$7 = $$4.e();
        ehg.a((ddc)$$0, $$1, (ayw)$$2, (jd)$$7, (efe)$$5);
        ehg.a((ddc)$$0, $$1, (ayw)$$2, (jd)$$7.i(), (efe)$$5);
        ehg.a((ddc)$$0, $$1, (ayw)$$2, (jd)$$7.g(), (efe)$$5);
        ehg.a((ddc)$$0, $$1, (ayw)$$2, (jd)$$7.g().i(), (efe)$$5);
        ji $$8 = ji.c.a.a($$2);
        int $$9 = $$3 - $$2.a(4);
        int $$10 = 2 - $$2.a(3);
        int $$11 = $$4.u();
        int $$12 = $$4.v();
        int $$13 = $$4.w();
        int $$14 = $$11;
        int $$15 = $$13;
        int $$16 = $$12 + $$3 - 1;
        for (int $$17 = 0; $$17 < $$3; ++$$17) {
            int $$18;
            jd $$19;
            if ($$17 >= $$9 && $$10 > 0) {
                $$14 += $$8.j();
                $$15 += $$8.l();
                --$$10;
            }
            if (!edq.b((ddc)$$0, (jd)($$19 = new jd($$14, $$18 = $$12 + $$17, $$15)))) continue;
            this.b($$0, $$1, $$2, $$19, $$5);
            this.b($$0, $$1, $$2, $$19.i(), $$5);
            this.b($$0, $$1, $$2, $$19.g(), $$5);
            this.b($$0, $$1, $$2, $$19.i().g(), $$5);
        }
        $$6.add(new efu.a(new jd($$14, $$16, $$15), 0, true));
        for (int $$20 = -1; $$20 <= 2; ++$$20) {
            for (int $$21 = -1; $$21 <= 2; ++$$21) {
                if ($$20 >= 0 && $$20 <= 1 && $$21 >= 0 && $$21 <= 1 || $$2.a(3) > 0) continue;
                int $$22 = $$2.a(3) + 2;
                for (int $$23 = 0; $$23 < $$22; ++$$23) {
                    this.b($$0, $$1, $$2, new jd($$11 + $$20, $$16 - $$23 - 1, $$13 + $$21), $$5);
                }
                $$6.add(new efu.a(new jd($$14 + $$20, $$16, $$15 + $$21), 0, false));
            }
        }
        return $$6;
    }
}
