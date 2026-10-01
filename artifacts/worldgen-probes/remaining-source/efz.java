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

public class efz
extends efu {
    public static final MapCodec<efz> a = RecordCodecBuilder.mapCodec($$02 -> efz.b($$02).and($$02.group((App)bpw.b(1, 512).fieldOf("foliage_height").forGetter($$0 -> $$0.b), (App)Codec.intRange((int)0, (int)256).fieldOf("leaf_placement_attempts").forGetter($$0 -> $$0.c))).apply((Applicative)$$02, efz::new));
    private final bpw b;
    private final int c;

    public efz(bpw $$0, bpw $$1, bpw $$2, int $$3) {
        super($$0, $$1);
        this.b = $$2;
        this.c = $$3;
    }

    @Override
    protected efv<?> a() {
        return efv.j;
    }

    @Override
    protected void a(ddc $$0, efu.b $$1, ayw $$2, efe $$3, int $$4, efu.a $$5, int $$6, int $$7, int $$8) {
        jd $$9 = $$5.a();
        jd.a $$10 = $$9.k();
        for (int $$11 = 0; $$11 < this.c; ++$$11) {
            $$10.a($$9, $$2.a($$7) - $$2.a($$7), $$2.a($$6) - $$2.a($$6), $$2.a($$7) - $$2.a($$7));
            efz.a($$0, $$1, $$2, $$3, $$10);
        }
    }

    @Override
    public int a(ayw $$0, int $$1, efe $$2) {
        return this.b.a($$0);
    }

    @Override
    protected boolean a(ayw $$0, int $$1, int $$2, int $$3, int $$4, boolean $$5) {
        return false;
    }
}
