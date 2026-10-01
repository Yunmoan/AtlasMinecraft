/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.datafixers.kinds.App
 *  com.mojang.datafixers.kinds.Applicative
 *  com.mojang.serialization.Codec
 *  com.mojang.serialization.codecs.RecordCodecBuilder
 */
import com.mojang.datafixers.kinds.App;
import com.mojang.datafixers.kinds.Applicative;
import com.mojang.serialization.Codec;
import com.mojang.serialization.codecs.RecordCodecBuilder;
import java.util.Optional;

public class eeh
implements eei {
    public static final Codec<eeh> a = RecordCodecBuilder.create($$02 -> $$02.group((App)jd.a.optionalFieldOf("exit").forGetter($$0 -> $$0.b), (App)Codec.BOOL.fieldOf("exact").forGetter($$0 -> $$0.c)).apply((Applicative)$$02, eeh::new));
    private final Optional<jd> b;
    private final boolean c;

    private eeh(Optional<jd> $$0, boolean $$1) {
        this.b = $$0;
        this.c = $$1;
    }

    public static eeh a(jd $$0, boolean $$1) {
        return new eeh(Optional.of($$0), $$1);
    }

    public static eeh a() {
        return new eeh(Optional.empty(), false);
    }

    public Optional<jd> b() {
        return this.b;
    }

    public boolean c() {
        return this.c;
    }
}
