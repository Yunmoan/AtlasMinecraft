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
import java.util.List;

public class bqd
extends bpw {
    public static final MapCodec<bqd> a = RecordCodecBuilder.mapCodec($$02 -> $$02.group((App)bpb.b(bpw.c).fieldOf("distribution").forGetter($$0 -> $$0.b)).apply((Applicative)$$02, bqd::new));
    private final bpb<bpw> b;
    private final int f;
    private final int g;

    public bqd(bpb<bpw> $$0) {
        this.b = $$0;
        List $$1 = $$0.e();
        int $$2 = Integer.MAX_VALUE;
        int $$3 = Integer.MIN_VALUE;
        for (bpd.b $$4 : $$1) {
            int $$5 = ((bpw)$$4.b()).a();
            int $$6 = ((bpw)$$4.b()).b();
            $$2 = Math.min($$2, $$5);
            $$3 = Math.max($$3, $$6);
        }
        this.f = $$2;
        this.g = $$3;
    }

    @Override
    public int a(ayw $$0) {
        return this.b.a($$0).orElseThrow(IllegalStateException::new).a($$0);
    }

    @Override
    public int a() {
        return this.f;
    }

    @Override
    public int b() {
        return this.g;
    }

    @Override
    public bpx<?> c() {
        return bpx.e;
    }
}
