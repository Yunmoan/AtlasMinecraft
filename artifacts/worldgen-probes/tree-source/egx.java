/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.serialization.Codec
 *  com.mojang.serialization.MapCodec
 *  it.unimi.dsi.fastutil.objects.ObjectArrayList
 */
import com.mojang.serialization.Codec;
import com.mojang.serialization.MapCodec;
import it.unimi.dsi.fastutil.objects.ObjectArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;
import java.util.stream.Stream;

public class egx
extends eha {
    public static final MapCodec<egx> a = Codec.floatRange((float)0.0f, (float)1.0f).fieldOf("probability").xmap(egx::new, $$0 -> Float.valueOf($$0.d));
    private static final ji b = ji.d;
    private static final ji[] c = (ji[])ji.c.a.a().filter($$0 -> $$0 != b.g()).toArray(ji[]::new);
    private final float d;

    public egx(float $$0) {
        this.d = $$0;
    }

    @Override
    protected ehb<?> a() {
        return ehb.d;
    }

    @Override
    public void a(eha.a $$02) {
        ayw $$12 = $$02.b();
        if ($$12.i() >= this.d) {
            return;
        }
        ObjectArrayList<jd> $$2 = $$02.d();
        ObjectArrayList<jd> $$3 = $$02.c();
        int $$4 = !$$2.isEmpty() ? Math.max(((jd)$$2.get(0)).v() - 1, ((jd)$$3.get(0)).v() + 1) : Math.min(((jd)$$3.get(0)).v() + 1 + $$12.a(3), ((jd)$$3.get($$3.size() - 1)).v());
        List $$5 = $$3.stream().filter($$1 -> $$1.v() == $$4).flatMap($$0 -> Stream.of(c).map($$0::a)).collect(Collectors.toList());
        if ($$5.isEmpty()) {
            return;
        }
        Collections.shuffle($$5);
        Optional<jd> $$6 = $$5.stream().filter($$1 -> $$02.a((jd)$$1) && $$02.a($$1.a(b))).findFirst();
        if ($$6.isEmpty()) {
            return;
        }
        $$02.a($$6.get(), (dtc)dga.pe.o().a(dfs.b, b));
        $$02.a().a($$6.get(), dqj.H).ifPresent($$1 -> {
            int $$2 = 2 + $$12.a(2);
            for (int $$3 = 0; $$3 < $$2; ++$$3) {
                $$1.a(dqe.c.a($$12.a(599)));
            }
        });
    }
}
