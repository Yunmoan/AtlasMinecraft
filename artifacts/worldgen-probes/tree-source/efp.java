/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.datafixers.Products$P3
 *  com.mojang.datafixers.kinds.App
 *  com.mojang.datafixers.kinds.Applicative
 *  com.mojang.serialization.Codec
 *  com.mojang.serialization.MapCodec
 *  com.mojang.serialization.codecs.RecordCodecBuilder
 *  com.mojang.serialization.codecs.RecordCodecBuilder$Instance
 *  com.mojang.serialization.codecs.RecordCodecBuilder$Mu
 */
import com.mojang.datafixers.Products;
import com.mojang.datafixers.kinds.App;
import com.mojang.datafixers.kinds.Applicative;
import com.mojang.serialization.Codec;
import com.mojang.serialization.MapCodec;
import com.mojang.serialization.codecs.RecordCodecBuilder;

public class efp
extends efu {
    public static final MapCodec<efp> a = RecordCodecBuilder.mapCodec($$0 -> efp.a($$0).apply((Applicative)$$0, efp::new));
    protected final int b;

    protected static <P extends efp> Products.P3<RecordCodecBuilder.Mu<P>, bpw, bpw, Integer> a(RecordCodecBuilder.Instance<P> $$02) {
        return efp.b($$02).and((App)Codec.intRange((int)0, (int)16).fieldOf("height").forGetter($$0 -> $$0.b));
    }

    public efp(bpw $$0, bpw $$1, int $$2) {
        super($$0, $$1);
        this.b = $$2;
    }

    @Override
    protected efv<?> a() {
        return efv.a;
    }

    @Override
    protected void a(ddc $$0, efu.b $$1, ayw $$2, efe $$3, int $$4, efu.a $$5, int $$6, int $$7, int $$8) {
        for (int $$9 = $$8; $$9 >= $$8 - $$6; --$$9) {
            int $$10 = Math.max($$7 + $$5.b() - 1 - $$9 / 2, 0);
            this.a($$0, $$1, $$2, $$3, $$5.a(), $$10, $$9, $$5.c());
        }
    }

    @Override
    public int a(ayw $$0, int $$1, efe $$2) {
        return this.b;
    }

    @Override
    protected boolean a(ayw $$0, int $$1, int $$2, int $$3, int $$4, boolean $$5) {
        return $$1 == $$4 && $$3 == $$4 && ($$0.a(2) == 0 || $$2 == 0);
    }
}
