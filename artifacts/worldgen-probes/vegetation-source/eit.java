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

public class eit
extends ejc {
    public static final MapCodec<eit> a = RecordCodecBuilder.mapCodec($$02 -> $$02.group((App)Codec.INT.fieldOf("noise_to_count_ratio").forGetter($$0 -> $$0.c), (App)Codec.DOUBLE.fieldOf("noise_factor").forGetter($$0 -> $$0.d), (App)Codec.DOUBLE.fieldOf("noise_offset").orElse((Object)0.0).forGetter($$0 -> $$0.e)).apply((Applicative)$$02, eit::new));
    private final int c;
    private final double d;
    private final double e;

    private eit(int $$0, double $$1, double $$2) {
        this.c = $$0;
        this.d = $$1;
        this.e = $$2;
    }

    public static eit a(int $$0, double $$1, double $$2) {
        return new eit($$0, $$1, $$2);
    }

    @Override
    protected int a(ayw $$0, jd $$1) {
        double $$2 = ddw.e.a((double)$$1.u() / this.d, (double)$$1.w() / this.d, false);
        return (int)Math.ceil(($$2 + this.e) * (double)this.c);
    }

    @Override
    public eiz<?> b() {
        return eiz.g;
    }
}

