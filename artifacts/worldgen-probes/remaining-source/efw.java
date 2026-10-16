/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.datafixers.kinds.App
 *  com.mojang.datafixers.kinds.Applicative
 *  com.mojang.serialization.Codec
 *  com.mojang.serialization.MapCodec
 *  com.mojang.serialization.codecs.RecordCodecBuilder
 */
import com.mojang.datafixers.kinds.App;
import com.mojang.datafixers.kinds.Applicative;
import com.mojang.serialization.Codec;
import com.mojang.serialization.MapCodec;
import com.mojang.serialization.codecs.RecordCodecBuilder;

public class efw
extends efu {
    public static final MapCodec<efw> a = RecordCodecBuilder.mapCodec($$02 -> efw.b($$02).and((App)Codec.intRange((int)0, (int)16).fieldOf("height").forGetter($$0 -> $$0.b)).apply((Applicative)$$02, efw::new));
    protected final int b;

    public efw(bpw $$0, bpw $$1, int $$2) {
        super($$0, $$1);
        this.b = $$2;
    }

    @Override
    protected efv<?> a() {
        return efv.g;
    }

    @Override
    protected void a(ddc $$0, efu.b $$1, ayw $$2, efe $$3, int $$4, efu.a $$5, int $$6, int $$7, int $$8) {
        int $$9 = $$5.c() ? $$6 : 1 + $$2.a(2);
        for (int $$10 = $$8; $$10 >= $$8 - $$9; --$$10) {
            int $$11 = $$7 + $$5.b() + 1 - $$10;
            this.a($$0, $$1, $$2, $$3, $$5.a(), $$11, $$10, $$5.c());
        }
    }

    @Override
    public int a(ayw $$0, int $$1, efe $$2) {
        return this.b;
    }

    @Override
    protected boolean a(ayw $$0, int $$1, int $$2, int $$3, int $$4, boolean $$5) {
        if ($$1 + $$3 >= 7) {
            return true;
        }
        return $$1 * $$1 + $$3 * $$3 > $$4 * $$4;
    }
}
