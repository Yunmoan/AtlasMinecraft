/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.datafixers.kinds.App
 *  com.mojang.datafixers.kinds.Applicative
 *  com.mojang.serialization.MapCodec
 *  com.mojang.serialization.codecs.RecordCodecBuilder
 */
import com.mojang.datafixers.kinds.App;
import com.mojang.datafixers.kinds.Applicative;
import com.mojang.serialization.MapCodec;
import com.mojang.serialization.codecs.RecordCodecBuilder;

public class efy
extends efu {
    public static final MapCodec<efy> a = RecordCodecBuilder.mapCodec($$02 -> efy.b($$02).and((App)bpw.b(0, 24).fieldOf("height").forGetter($$0 -> $$0.b)).apply((Applicative)$$02, efy::new));
    private final bpw b;

    public efy(bpw $$0, bpw $$1, bpw $$2) {
        super($$0, $$1);
        this.b = $$2;
    }

    @Override
    protected efv<?> a() {
        return efv.c;
    }

    @Override
    protected void a(ddc $$0, efu.b $$1, ayw $$2, efe $$3, int $$4, efu.a $$5, int $$6, int $$7, int $$8) {
        int $$9 = 0;
        for (int $$10 = $$8; $$10 >= $$8 - $$6; --$$10) {
            this.a($$0, $$1, $$2, $$3, $$5.a(), $$9, $$10, $$5.c());
            if ($$9 >= 1 && $$10 == $$8 - $$6 + 1) {
                --$$9;
                continue;
            }
            if ($$9 >= $$7 + $$5.b()) continue;
            ++$$9;
        }
    }

    @Override
    public int a(ayw $$0, int $$1) {
        return super.a($$0, $$1) + $$0.a(Math.max($$1 + 1, 1));
    }

    @Override
    public int a(ayw $$0, int $$1, efe $$2) {
        return this.b.a($$0);
    }

    @Override
    protected boolean a(ayw $$0, int $$1, int $$2, int $$3, int $$4, boolean $$5) {
        return $$1 == $$4 && $$3 == $$4 && $$4 > 0;
    }
}
