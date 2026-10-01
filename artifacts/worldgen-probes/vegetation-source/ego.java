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
import java.util.List;

public class ego
extends egm {
    public static final MapCodec<ego> b = RecordCodecBuilder.mapCodec($$02 -> ego.a($$02).and($$02.group((App)Codec.floatRange((float)-1.0f, (float)1.0f).fieldOf("threshold").forGetter($$0 -> Float.valueOf($$0.g)), (App)Codec.floatRange((float)0.0f, (float)1.0f).fieldOf("high_chance").forGetter($$0 -> Float.valueOf($$0.h)), (App)dtc.b.fieldOf("default_state").forGetter($$0 -> $$0.i), (App)Codec.list(dtc.b).fieldOf("low_states").forGetter($$0 -> $$0.j), (App)Codec.list(dtc.b).fieldOf("high_states").forGetter($$0 -> $$0.k))).apply((Applicative)$$02, ego::new));
    private final float g;
    private final float h;
    private final dtc i;
    private final List<dtc> j;
    private final List<dtc> k;

    public ego(long $$0, eoh.a $$1, float $$2, float $$3, float $$4, dtc $$5, List<dtc> $$6, List<dtc> $$7) {
        super($$0, $$1, $$2);
        this.g = $$3;
        this.h = $$4;
        this.i = $$5;
        this.j = $$6;
        this.k = $$7;
    }

    @Override
    protected egk<?> a() {
        return egk.c;
    }

    @Override
    public dtc a(ayw $$0, jd $$1) {
        double $$2 = this.a($$1, this.e);
        if ($$2 < (double)this.g) {
            return ad.a(this.j, $$0);
        }
        if ($$0.i() < this.h) {
            return ad.a(this.k, $$0);
        }
        return this.i;
    }
}

