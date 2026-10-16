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

public class egy
extends eha {
    public static final MapCodec<egy> a = Codec.floatRange((float)0.0f, (float)1.0f).fieldOf("probability").xmap(egy::new, $$0 -> Float.valueOf($$0.b));
    private final float b;

    public egy(float $$0) {
        this.b = $$0;
    }

    @Override
    protected ehb<?> a() {
        return ehb.c;
    }

    @Override
    public void a(eha.a $$0) {
        ayw $$12 = $$0.b();
        if ($$12.i() >= this.b) {
            return;
        }
        ObjectArrayList<jd> $$22 = $$0.c();
        int $$3 = ((jd)$$22.get(0)).v();
        $$22.stream().filter($$1 -> $$1.v() - $$3 <= 2).forEach($$2 -> {
            for (ji $$3 : ji.c.a) {
                ji $$4;
                jd $$5;
                if (!($$12.i() <= 0.25f) || !$$0.a($$5 = $$2.b(($$4 = $$3.g()).j(), 0, $$4.l()))) continue;
                $$0.a($$5, (dtc)((dtc)dga.fC.o().a(dhf.c, $$12.a(3))).a(dhf.aE, $$3));
            }
        });
    }
}
