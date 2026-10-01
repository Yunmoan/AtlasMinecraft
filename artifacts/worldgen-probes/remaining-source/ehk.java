/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  ayo
 *  ayw
 *  com.google.common.collect.Lists
 *  com.mojang.datafixers.kinds.Applicative
 *  com.mojang.serialization.MapCodec
 *  com.mojang.serialization.codecs.RecordCodecBuilder
 *  com.mojang.serialization.codecs.RecordCodecBuilder$Instance
 *  ddc
 *  dtc
 *  efe
 *  efu$a
 *  ehn
 *  jd
 */
import com.google.common.collect.Lists;
import com.mojang.datafixers.kinds.Applicative;
import com.mojang.serialization.MapCodec;
import com.mojang.serialization.codecs.RecordCodecBuilder;
import java.util.ArrayList;
import java.util.List;
import java.util.function.BiConsumer;

public class ehk
extends ehj {
    public static final MapCodec<ehk> b = RecordCodecBuilder.mapCodec($$0 -> ehk.a((RecordCodecBuilder.Instance)$$0).apply((Applicative)$$0, ehk::new));

    public ehk(int $$0, int $$1, int $$2) {
        super($$0, $$1, $$2);
    }

    @Override
    protected ehn<?> a() {
        return ehn.d;
    }

    @Override
    public List<efu.a> a(ddc $$0, BiConsumer<jd, dtc> $$1, ayw $$2, int $$3, jd $$4, efe $$5) {
        ArrayList $$6 = Lists.newArrayList();
        $$6.addAll(super.a($$0, $$1, $$2, $$3, $$4, $$5));
        for (int $$7 = $$3 - 2 - $$2.a(4); $$7 > $$3 / 2; $$7 -= 2 + $$2.a(4)) {
            float $$8 = $$2.i() * ((float)Math.PI * 2);
            int $$9 = 0;
            int $$10 = 0;
            for (int $$11 = 0; $$11 < 5; ++$$11) {
                $$9 = (int)(1.5f + ayo.b((float)$$8) * (float)$$11);
                $$10 = (int)(1.5f + ayo.a((float)$$8) * (float)$$11);
                jd $$12 = $$4.b($$9, $$7 - 3 + $$11 / 2, $$10);
                this.b($$0, $$1, $$2, $$12, $$5);
            }
            $$6.add(new efu.a($$4.b($$9, $$7, $$10), -2, false));
        }
        return $$6;
    }
}
