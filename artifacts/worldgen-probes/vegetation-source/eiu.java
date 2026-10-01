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

public class eiu
extends ejc {
    public static final MapCodec<eiu> a = RecordCodecBuilder.mapCodec($$02 -> $$02.group((App)Codec.DOUBLE.fieldOf("noise_level").forGetter($$0 -> $$0.c), (App)Codec.INT.fieldOf("below_noise").forGetter($$0 -> $$0.d), (App)Codec.INT.fieldOf("above_noise").forGetter($$0 -> $$0.e)).apply((Applicative)$$02, eiu::new));
    private final double c;
    private final int d;
    private final int e;

    private eiu(double $$0, int $$1, int $$2) {
        this.c = $$0;
        this.d = $$1;
        this.e = $$2;
    }

    public static eiu a(double $$0, int $$1, int $$2) {
        return new eiu($$0, $$1, $$2);
    }

    @Override
    protected int a(ayw $$0, jd $$1) {
        double $$2 = ddw.e.a((double)$$1.u() / 200.0, (double)$$1.w() / 200.0, false);
        return $$2 < this.c ? this.d : this.e;
    }

    @Override
    public eiz<?> b() {
        return eiz.h;
    }
}

