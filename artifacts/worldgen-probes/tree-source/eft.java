/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.datafixers.kinds.Applicative
 *  com.mojang.serialization.MapCodec
 *  com.mojang.serialization.codecs.RecordCodecBuilder
 */
import com.mojang.datafixers.kinds.Applicative;
import com.mojang.serialization.MapCodec;
import com.mojang.serialization.codecs.RecordCodecBuilder;

public class eft
extends efp {
    public static final MapCodec<eft> c = RecordCodecBuilder.mapCodec($$0 -> eft.a($$0).apply((Applicative)$$0, eft::new));

    public eft(bpw $$0, bpw $$1, int $$2) {
        super($$0, $$1, $$2);
    }

    @Override
    protected efv<?> a() {
        return efv.f;
    }

    @Override
    protected void a(ddc $$0, efu.b $$1, ayw $$2, efe $$3, int $$4, efu.a $$5, int $$6, int $$7, int $$8) {
        for (int $$9 = $$8; $$9 >= $$8 - $$6; --$$9) {
            int $$10 = $$7 + ($$9 == $$8 || $$9 == $$8 - $$6 ? 0 : 1);
            this.a($$0, $$1, $$2, $$3, $$5.a(), $$10, $$9, $$5.c());
        }
    }

    @Override
    protected boolean a(ayw $$0, int $$1, int $$2, int $$3, int $$4, boolean $$5) {
        return ayo.k((float)$$1 + 0.5f) + ayo.k((float)$$3 + 0.5f) > (float)($$4 * $$4);
    }
}
