/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.serialization.DataResult
 *  com.mojang.serialization.MapCodec
 */
import com.mojang.serialization.DataResult;
import com.mojang.serialization.MapCodec;

public class egt
extends egj {
    public static final MapCodec<egt> b = bpb.b(dtc.b).comapFlatMap(egt::a, $$0 -> $$0.c).fieldOf("entries");
    private final bpb<dtc> c;

    private static DataResult<egt> a(bpb<dtc> $$0) {
        if ($$0.d()) {
            return DataResult.error(() -> "WeightedStateProvider with no states");
        }
        return DataResult.success((Object)new egt($$0));
    }

    public egt(bpb<dtc> $$0) {
        this.c = $$0;
    }

    public egt(bpb.a<dtc> $$0) {
        this($$0.a());
    }

    @Override
    protected egk<?> a() {
        return egk.b;
    }

    @Override
    public dtc a(ayw $$0, jd $$1) {
        return this.c.a($$0).orElseThrow(IllegalStateException::new);
    }
}

