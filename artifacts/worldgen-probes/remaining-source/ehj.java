/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.google.common.collect.ImmutableList
 *  com.mojang.datafixers.kinds.Applicative
 *  com.mojang.serialization.MapCodec
 *  com.mojang.serialization.codecs.RecordCodecBuilder
 */
import com.google.common.collect.ImmutableList;
import com.mojang.datafixers.kinds.Applicative;
import com.mojang.serialization.MapCodec;
import com.mojang.serialization.codecs.RecordCodecBuilder;
import java.util.List;
import java.util.function.BiConsumer;

public class ehj
extends ehm {
    public static final MapCodec<ehj> a = RecordCodecBuilder.mapCodec($$0 -> ehj.a($$0).apply((Applicative)$$0, ehj::new));

    public ehj(int $$0, int $$1, int $$2) {
        super($$0, $$1, $$2);
    }

    @Override
    protected ehn<?> a() {
        return ehn.c;
    }

    @Override
    public List<efu.a> a(ddc $$0, BiConsumer<jd, dtc> $$1, ayw $$2, int $$3, jd $$4, efe $$5) {
        jd $$6 = $$4.e();
        ehj.a($$0, $$1, $$2, $$6, $$5);
        ehj.a($$0, $$1, $$2, $$6.i(), $$5);
        ehj.a($$0, $$1, $$2, $$6.g(), $$5);
        ehj.a($$0, $$1, $$2, $$6.g().i(), $$5);
        jd.a $$7 = new jd.a();
        for (int $$8 = 0; $$8 < $$3; ++$$8) {
            this.a($$0, $$1, $$2, $$7, $$5, $$4, 0, $$8, 0);
            if ($$8 >= $$3 - 1) continue;
            this.a($$0, $$1, $$2, $$7, $$5, $$4, 1, $$8, 0);
            this.a($$0, $$1, $$2, $$7, $$5, $$4, 1, $$8, 1);
            this.a($$0, $$1, $$2, $$7, $$5, $$4, 0, $$8, 1);
        }
        return ImmutableList.of((Object)new efu.a($$4.b($$3), 0, true));
    }

    private void a(ddc $$0, BiConsumer<jd, dtc> $$1, ayw $$2, jd.a $$3, efe $$4, jd $$5, int $$6, int $$7, int $$8) {
        $$3.a($$5, $$6, $$7, $$8);
        this.a($$0, $$1, $$2, $$3, $$4);
    }
}
