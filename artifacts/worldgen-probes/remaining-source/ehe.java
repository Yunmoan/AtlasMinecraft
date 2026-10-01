/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.google.common.collect.Lists
 *  com.mojang.datafixers.kinds.App
 *  com.mojang.datafixers.kinds.Applicative
 *  com.mojang.serialization.MapCodec
 *  com.mojang.serialization.codecs.RecordCodecBuilder
 */
import com.google.common.collect.Lists;
import com.mojang.datafixers.kinds.App;
import com.mojang.datafixers.kinds.Applicative;
import com.mojang.serialization.MapCodec;
import com.mojang.serialization.codecs.RecordCodecBuilder;
import java.util.ArrayList;
import java.util.List;
import java.util.function.BiConsumer;

public class ehe
extends ehm {
    public static final MapCodec<ehe> a = RecordCodecBuilder.mapCodec($$02 -> ehe.a($$02).and($$02.group((App)axw.l.optionalFieldOf("min_height_for_leaves", (Object)1).forGetter($$0 -> $$0.b), (App)bpw.b(1, 64).fieldOf("bend_length").forGetter($$0 -> $$0.h))).apply((Applicative)$$02, ehe::new));
    private final int b;
    private final bpw h;

    public ehe(int $$0, int $$1, int $$2, int $$3, bpw $$4) {
        super($$0, $$1, $$2);
        this.b = $$3;
        this.h = $$4;
    }

    @Override
    protected ehn<?> a() {
        return ehn.g;
    }

    @Override
    public List<efu.a> a(ddc $$0, BiConsumer<jd, dtc> $$1, ayw $$2, int $$3, jd $$4, efe $$5) {
        ji $$6 = ji.c.a.a($$2);
        int $$7 = $$3 - 1;
        jd.a $$8 = $$4.k();
        jd $$9 = $$8.e();
        ehe.a($$0, $$1, $$2, $$9, $$5);
        ArrayList $$10 = Lists.newArrayList();
        for (int $$11 = 0; $$11 <= $$7; ++$$11) {
            if ($$11 + 1 >= $$7 + $$2.a(2)) {
                $$8.c($$6);
            }
            if (edq.c($$0, $$8)) {
                this.b($$0, $$1, $$2, $$8, $$5);
            }
            if ($$11 >= this.b) {
                $$10.add(new efu.a($$8.j(), 0, false));
            }
            $$8.c(ji.b);
        }
        int $$12 = this.h.a($$2);
        for (int $$13 = 0; $$13 <= $$12; ++$$13) {
            if (edq.c($$0, $$8)) {
                this.b($$0, $$1, $$2, $$8, $$5);
            }
            $$10.add(new efu.a($$8.j(), 0, false));
            $$8.c($$6);
        }
        return $$10;
    }
}
