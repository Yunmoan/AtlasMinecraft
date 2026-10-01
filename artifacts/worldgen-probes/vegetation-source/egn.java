/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.datafixers.Products$P4
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
import java.util.List;

public class egn
extends egm {
    public static final MapCodec<egn> g = RecordCodecBuilder.mapCodec($$0 -> egn.b($$0).apply((Applicative)$$0, egn::new));
    protected final List<dtc> h;

    protected static <P extends egn> Products.P4<RecordCodecBuilder.Mu<P>, Long, eoh.a, Float, List<dtc>> b(RecordCodecBuilder.Instance<P> $$02) {
        return egn.a($$02).and((App)Codec.list(dtc.b).fieldOf("states").forGetter($$0 -> $$0.h));
    }

    public egn(long $$0, eoh.a $$1, float $$2, List<dtc> $$3) {
        super($$0, $$1, $$2);
        this.h = $$3;
    }

    @Override
    protected egk<?> a() {
        return egk.d;
    }

    @Override
    public dtc a(ayw $$0, jd $$1) {
        return this.a(this.h, $$1, this.e);
    }

    protected dtc a(List<dtc> $$0, jd $$1, double $$2) {
        double $$3 = this.a($$1, $$2);
        return this.a($$0, $$3);
    }

    protected dtc a(List<dtc> $$0, double $$1) {
        double $$2 = ayo.a((1.0 + $$1) / 2.0, 0.0, 0.9999);
        return $$0.get((int)($$2 * (double)$$0.size()));
    }
}

