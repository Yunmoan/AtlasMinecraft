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

public class eje
extends eix {
    public static final MapCodec<eje> a = RecordCodecBuilder.mapCodec($$02 -> $$02.group((App)Codec.INT.fieldOf("max_water_depth").forGetter($$0 -> $$0.c)).apply((Applicative)$$02, eje::new));
    private final int c;

    private eje(int $$0) {
        this.c = $$0;
    }

    public static eje a(int $$0) {
        return new eje($$0);
    }

    @Override
    protected boolean a(eiw $$0, ayw $$1, jd $$2) {
        int $$3 = $$0.a(dyy.a.d, $$2.u(), $$2.w());
        int $$4 = $$0.a(dyy.a.b, $$2.u(), $$2.w());
        return $$4 - $$3 <= this.c;
    }

    @Override
    public eiz<?> b() {
        return eiz.d;
    }
}
