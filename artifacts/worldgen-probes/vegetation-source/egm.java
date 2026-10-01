/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.datafixers.Products$P3
 *  com.mojang.datafixers.kinds.App
 *  com.mojang.serialization.Codec
 *  com.mojang.serialization.codecs.RecordCodecBuilder$Instance
 *  com.mojang.serialization.codecs.RecordCodecBuilder$Mu
 */
import com.mojang.datafixers.Products;
import com.mojang.datafixers.kinds.App;
import com.mojang.serialization.Codec;
import com.mojang.serialization.codecs.RecordCodecBuilder;

public abstract class egm
extends egj {
    protected final long c;
    protected final eoh.a d;
    protected final float e;
    protected final eoh f;

    protected static <P extends egm> Products.P3<RecordCodecBuilder.Mu<P>, Long, eoh.a, Float> a(RecordCodecBuilder.Instance<P> $$02) {
        return $$02.group((App)Codec.LONG.fieldOf("seed").forGetter($$0 -> $$0.c), (App)eoh.a.a.fieldOf("noise").forGetter($$0 -> $$0.d), (App)axw.m.fieldOf("scale").forGetter($$0 -> Float.valueOf($$0.e)));
    }

    protected egm(long $$0, eoh.a $$1, float $$2) {
        this.c = $$0;
        this.d = $$1;
        this.e = $$2;
        this.f = eoh.b(new dzx(new dyz($$0)), $$1);
    }

    protected double a(jd $$0, double $$1) {
        return this.f.a((double)$$0.u() * $$1, (double)$$0.v() * $$1, (double)$$0.w() * $$1);
    }
}

