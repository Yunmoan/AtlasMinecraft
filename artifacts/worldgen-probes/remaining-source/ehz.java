/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  ayo
 *  ayw
 *  com.mojang.datafixers.kinds.App
 *  com.mojang.datafixers.kinds.Applicative
 *  com.mojang.logging.LogUtils
 *  com.mojang.serialization.Codec
 *  com.mojang.serialization.MapCodec
 *  com.mojang.serialization.codecs.RecordCodecBuilder
 *  dzs
 *  dzv
 *  ehx
 *  ehy
 *  org.slf4j.Logger
 */
import com.mojang.datafixers.kinds.App;
import com.mojang.datafixers.kinds.Applicative;
import com.mojang.logging.LogUtils;
import com.mojang.serialization.Codec;
import com.mojang.serialization.MapCodec;
import com.mojang.serialization.codecs.RecordCodecBuilder;
import org.slf4j.Logger;

public class ehz
extends ehx {
    public static final MapCodec<ehz> a = RecordCodecBuilder.mapCodec($$02 -> $$02.group((App)dzs.a.fieldOf("min_inclusive").forGetter($$0 -> $$0.d), (App)dzs.a.fieldOf("max_inclusive").forGetter($$0 -> $$0.e), (App)Codec.INT.optionalFieldOf("plateau", (Object)0).forGetter($$0 -> $$0.f)).apply((Applicative)$$02, ehz::new));
    private static final Logger b = LogUtils.getLogger();
    private final dzs d;
    private final dzs e;
    private final int f;

    private ehz(dzs $$0, dzs $$1, int $$2) {
        this.d = $$0;
        this.e = $$1;
        this.f = $$2;
    }

    public static ehz a(dzs $$0, dzs $$1, int $$2) {
        return new ehz($$0, $$1, $$2);
    }

    public static ehz a(dzs $$0, dzs $$1) {
        return ehz.a($$0, $$1, 0);
    }

    public int a(ayw $$0, dzv $$1) {
        int $$3;
        int $$2 = this.d.a($$1);
        if ($$2 > ($$3 = this.e.a($$1))) {
            b.warn("Empty height range: {}", (Object)this);
            return $$2;
        }
        int $$4 = $$3 - $$2;
        if (this.f >= $$4) {
            return ayo.b((ayw)$$0, (int)$$2, (int)$$3);
        }
        int $$5 = ($$4 - this.f) / 2;
        int $$6 = $$4 - $$5;
        return $$2 + ayo.b((ayw)$$0, (int)0, (int)$$6) + ayo.b((ayw)$$0, (int)0, (int)$$5);
    }

    public ehy<?> a() {
        return ehy.e;
    }

    public String toString() {
        if (this.f == 0) {
            return "triangle (" + String.valueOf(this.d) + "-" + String.valueOf(this.e) + ")";
        }
        return "trapezoid(" + this.f + ") in [" + String.valueOf(this.d) + "-" + String.valueOf(this.e) + "]";
    }
}
