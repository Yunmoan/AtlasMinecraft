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

public class ehl
extends ehm {
    public static final MapCodec<ehl> a = RecordCodecBuilder.mapCodec($$0 -> ehl.a($$0).apply((Applicative)$$0, ehl::new));

    public ehl(int $$0, int $$1, int $$2) {
        super($$0, $$1, $$2);
    }

    @Override
    protected ehn<?> a() {
        return ehn.a;
    }

    @Override
    public List<efu.a> a(ddc $$0, BiConsumer<jd, dtc> $$1, ayw $$2, int $$3, jd $$4, efe $$5) {
        ehl.a($$0, $$1, $$2, $$4.e(), $$5);
        for (int $$6 = 0; $$6 < $$3; ++$$6) {
            this.b($$0, $$1, $$2, $$4.b($$6), $$5);
        }
        return ImmutableList.of((Object)new efu.a($$4.b($$3), 0, false));
    }
}
