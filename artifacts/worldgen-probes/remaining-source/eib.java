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

public class eib
extends ehx {
    public static final MapCodec<eib> a = RecordCodecBuilder.mapCodec($$02 -> $$02.group((App)dzs.a.fieldOf("min_inclusive").forGetter($$0 -> $$0.d), (App)dzs.a.fieldOf("max_inclusive").forGetter($$0 -> $$0.e), (App)Codec.intRange((int)1, (int)Integer.MAX_VALUE).optionalFieldOf("inner", (Object)1).forGetter($$0 -> $$0.f)).apply((Applicative)$$02, eib::new));
    private static final Logger b = LogUtils.getLogger();
    private final dzs d;
    private final dzs e;
    private final int f;

    private eib(dzs $$0, dzs $$1, int $$2) {
        this.d = $$0;
        this.e = $$1;
        this.f = $$2;
    }

    public static eib a(dzs $$0, dzs $$1, int $$2) {
        return new eib($$0, $$1, $$2);
    }

    public int a(ayw $$0, dzv $$1) {
        int $$2 = this.d.a($$1);
        int $$3 = this.e.a($$1);
        if ($$3 - $$2 - this.f + 1 <= 0) {
            b.warn("Empty height range: {}", (Object)this);
            return $$2;
        }
        int $$4 = ayo.a((ayw)$$0, (int)($$2 + this.f), (int)$$3);
        int $$5 = ayo.a((ayw)$$0, (int)$$2, (int)($$4 - 1));
        return ayo.a((ayw)$$0, (int)$$2, (int)($$5 - 1 + this.f));
    }

    public ehy<?> a() {
        return ehy.d;
    }

    public String toString() {
        return "biased[" + String.valueOf(this.d) + "-" + String.valueOf(this.e) + " inner: " + this.f + "]";
    }
}
