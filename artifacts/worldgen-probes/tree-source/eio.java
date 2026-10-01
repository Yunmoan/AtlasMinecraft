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
import java.util.stream.Stream;

public class eio
extends eiy {
    private final ji c;
    private final eaf d;
    private final eaf e;
    private final int f;
    public static final MapCodec<eio> a = RecordCodecBuilder.mapCodec($$02 -> $$02.group((App)ji.h.fieldOf("direction_of_search").forGetter($$0 -> $$0.c), (App)eaf.b.fieldOf("target_condition").forGetter($$0 -> $$0.d), (App)eaf.b.optionalFieldOf("allowed_search_condition", (Object)eaf.e()).forGetter($$0 -> $$0.e), (App)Codec.intRange((int)1, (int)32).fieldOf("max_steps").forGetter($$0 -> $$0.f)).apply((Applicative)$$02, eio::new));

    private eio(ji $$0, eaf $$1, eaf $$2, int $$3) {
        this.c = $$0;
        this.d = $$1;
        this.e = $$2;
        this.f = $$3;
    }

    public static eio a(ji $$0, eaf $$1, eaf $$2, int $$3) {
        return new eio($$0, $$1, $$2, $$3);
    }

    public static eio a(ji $$0, eaf $$1, int $$2) {
        return eio.a($$0, $$1, eaf.e(), $$2);
    }

    @Override
    public Stream<jd> a_(eiw $$0, ayw $$1, jd $$2) {
        jd.a $$3 = $$2.k();
        dds $$4 = $$0.d();
        if (!this.e.test($$4, $$3)) {
            return Stream.of(new jd[0]);
        }
        for (int $$5 = 0; $$5 < this.f; ++$$5) {
            if (this.d.test($$4, $$3)) {
                return Stream.of($$3);
            }
            $$3.c(this.c);
            if ($$4.d($$3.v())) {
                return Stream.of(new jd[0]);
            }
            if (!this.e.test($$4, $$3)) break;
        }
        if (this.d.test($$4, $$3)) {
            return Stream.of($$3);
        }
        return Stream.of(new jd[0]);
    }

    @Override
    public eiz<?> b() {
        return eiz.j;
    }
}
