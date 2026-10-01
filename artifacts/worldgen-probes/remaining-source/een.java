/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.datafixers.kinds.App
 *  com.mojang.datafixers.kinds.Applicative
 *  com.mojang.serialization.Codec
 *  com.mojang.serialization.DataResult
 *  com.mojang.serialization.codecs.RecordCodecBuilder
 *  it.unimi.dsi.fastutil.objects.ObjectArrayList
 */
import com.mojang.datafixers.kinds.App;
import com.mojang.datafixers.kinds.Applicative;
import com.mojang.serialization.Codec;
import com.mojang.serialization.DataResult;
import com.mojang.serialization.codecs.RecordCodecBuilder;
import it.unimi.dsi.fastutil.objects.ObjectArrayList;
import java.util.List;

public class een
implements eei {
    public static final Codec<een> a = RecordCodecBuilder.create($$02 -> $$02.group((App)lt.e.r().fieldOf("block").flatXmap(een::a, DataResult::success).orElse((Object)((dky)dga.fg)).forGetter($$0 -> $$0.b), (App)Codec.intRange((int)1, (int)64).fieldOf("search_range").orElse((Object)10).forGetter($$0 -> $$0.c), (App)Codec.BOOL.fieldOf("can_place_on_floor").orElse((Object)false).forGetter($$0 -> $$0.d), (App)Codec.BOOL.fieldOf("can_place_on_ceiling").orElse((Object)false).forGetter($$0 -> $$0.e), (App)Codec.BOOL.fieldOf("can_place_on_wall").orElse((Object)false).forGetter($$0 -> $$0.f), (App)Codec.floatRange((float)0.0f, (float)1.0f).fieldOf("chance_of_spreading").orElse((Object)Float.valueOf(0.5f)).forGetter($$0 -> Float.valueOf($$0.g)), (App)kb.a(lu.f).fieldOf("can_be_placed_on").forGetter($$0 -> $$0.h)).apply((Applicative)$$02, een::new));
    public final dky b;
    public final int c;
    public final boolean d;
    public final boolean e;
    public final boolean f;
    public final float g;
    public final jq<dfy> h;
    private final ObjectArrayList<ji> i;

    private static DataResult<dky> a(dfy $$0) {
        DataResult dataResult;
        if ($$0 instanceof dky) {
            dky $$1 = (dky)$$0;
            dataResult = DataResult.success((Object)$$1);
        } else {
            dataResult = DataResult.error(() -> "Growth block should be a multiface block");
        }
        return dataResult;
    }

    public een(dky $$0, int $$1, boolean $$2, boolean $$3, boolean $$4, float $$5, jq<dfy> $$6) {
        this.b = $$0;
        this.c = $$1;
        this.d = $$2;
        this.e = $$3;
        this.f = $$4;
        this.g = $$5;
        this.h = $$6;
        this.i = new ObjectArrayList(6);
        if ($$3) {
            this.i.add((Object)ji.b);
        }
        if ($$2) {
            this.i.add((Object)ji.a);
        }
        if ($$4) {
            ji.c.a.forEach(arg_0 -> this.i.add(arg_0));
        }
    }

    public List<ji> a(ayw $$0, ji $$12) {
        return ad.a(this.i.stream().filter($$1 -> $$1 != $$12), $$0);
    }

    public List<ji> a(ayw $$0) {
        return ad.a(this.i, $$0);
    }
}
