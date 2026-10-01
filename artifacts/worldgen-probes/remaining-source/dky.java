/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.google.common.collect.ImmutableMap
 *  com.google.common.collect.Maps
 *  com.mojang.serialization.MapCodec
 *  javax.annotation.Nullable
 */
import com.google.common.collect.ImmutableMap;
import com.google.common.collect.Maps;
import com.mojang.serialization.MapCodec;
import java.util.Arrays;
import java.util.Collection;
import java.util.EnumSet;
import java.util.Map;
import java.util.Objects;
import java.util.Set;
import java.util.function.Function;
import javax.annotation.Nullable;

public abstract class dky
extends dfy {
    private static final float a = 1.0f;
    private static final exv c = dfy.a(0.0, 15.0, 0.0, 16.0, 16.0, 16.0);
    private static final exv d = dfy.a(0.0, 0.0, 0.0, 16.0, 1.0, 16.0);
    private static final exv e = dfy.a(0.0, 0.0, 0.0, 1.0, 16.0, 16.0);
    private static final exv f = dfy.a(15.0, 0.0, 0.0, 16.0, 16.0, 16.0);
    private static final exv g = dfy.a(0.0, 0.0, 0.0, 16.0, 16.0, 1.0);
    private static final exv h = dfy.a(0.0, 0.0, 15.0, 16.0, 16.0, 16.0);
    private static final Map<ji, dtt> i = dlm.h;
    private static final Map<ji, exv> j = ad.a(Maps.newEnumMap(ji.class), (? super T $$0) -> {
        $$0.put(ji.c, g);
        $$0.put(ji.f, f);
        $$0.put(ji.d, h);
        $$0.put(ji.e, e);
        $$0.put(ji.b, c);
        $$0.put(ji.a, d);
    });
    protected static final ji[] b = ji.values();
    private final ImmutableMap<dtc, exv> k;
    private final boolean l;
    private final boolean m;
    private final boolean n;

    public dky(dtb.d $$0) {
        super($$0);
        this.k(dky.a(this.E));
        this.k = this.a(dky::o);
        this.l = ji.c.a.a().allMatch(this::a);
        this.m = ji.c.a.a().filter(ji.a.a).filter(this::a).count() % 2L == 0L;
        this.n = ji.c.a.a().filter(ji.a.c).filter(this::a).count() % 2L == 0L;
    }

    protected abstract MapCodec<? extends dky> a();

    public static Set<ji> m(dtc $$0) {
        if (!($$0.b() instanceof dky)) {
            return Set.of();
        }
        EnumSet<ji> $$1 = EnumSet.noneOf(ji.class);
        for (ji $$2 : ji.values()) {
            if (!dky.a($$0, $$2)) continue;
            $$1.add($$2);
        }
        return $$1;
    }

    public static Set<ji> a(byte $$0) {
        EnumSet<ji> $$1 = EnumSet.noneOf(ji.class);
        for (ji $$2 : ji.values()) {
            if (($$0 & (byte)(1 << $$2.ordinal())) <= 0) continue;
            $$1.add($$2);
        }
        return $$1;
    }

    public static byte a(Collection<ji> $$0) {
        byte $$1 = 0;
        for (ji $$2 : $$0) {
            $$1 = (byte)($$1 | 1 << $$2.ordinal());
        }
        return $$1;
    }

    protected boolean a(ji $$0) {
        return true;
    }

    @Override
    protected void a(dtd.a<dfy, dtc> $$0) {
        for (ji $$1 : b) {
            if (!this.a($$1)) continue;
            $$0.a(new duf[]{dky.b($$1)});
        }
    }

    @Override
    protected dtc a(dtc $$0, ji $$1, dtc $$2, dcx $$3, jd $$4, jd $$5) {
        if (!dky.n($$0)) {
            return dga.a.o();
        }
        if (!dky.a($$0, $$1) || dky.a($$3, $$1, $$5, $$2)) {
            return $$0;
        }
        return dky.a($$0, dky.b($$1));
    }

    @Override
    protected exv a(dtc $$0, dcc $$1, jd $$2, exh $$3) {
        return (exv)this.k.get((Object)$$0);
    }

    @Override
    protected boolean a(dtc $$0, dcz $$1, jd $$2) {
        boolean $$3 = false;
        for (ji $$4 : b) {
            if (!dky.a($$0, $$4)) continue;
            jd $$5 = $$2.a($$4);
            if (!dky.a($$1, $$4, $$5, $$1.a_($$5))) {
                return false;
            }
            $$3 = true;
        }
        return $$3;
    }

    @Override
    protected boolean a(dtc $$0, cyd $$1) {
        return dky.p($$0);
    }

    @Override
    @Nullable
    public dtc a(cyd $$0) {
        dcw $$1 = $$0.q();
        jd $$2 = $$0.a();
        dtc $$32 = $$1.a_($$2);
        return Arrays.stream($$0.f()).map($$3 -> this.c($$32, (dcc)$$1, $$2, (ji)$$3)).filter(Objects::nonNull).findFirst().orElse(null);
    }

    public boolean a(dcc $$0, dtc $$1, jd $$2, ji $$3) {
        if (!this.a($$3) || $$1.a(this) && dky.a($$1, $$3)) {
            return false;
        }
        jd $$4 = $$2.a($$3);
        return dky.a($$0, $$3, $$4, $$0.a_($$4));
    }

    @Nullable
    public dtc c(dtc $$0, dcc $$1, jd $$2, ji $$3) {
        dtc $$6;
        if (!this.a($$1, $$0, $$2, $$3)) {
            return null;
        }
        if ($$0.a(this)) {
            dtc $$4 = $$0;
        } else if (this.m() && $$0.u().a(epf.c)) {
            dtc $$5 = (dtc)this.o().a(dts.C, true);
        } else {
            $$6 = this.o();
        }
        return (dtc)$$6.a(dky.b($$3), true);
    }

    @Override
    protected dtc a(dtc $$0, dmm $$1) {
        if (!this.l) {
            return $$0;
        }
        return this.a($$0, $$1::a);
    }

    @Override
    protected dtc a(dtc $$0, dkv $$1) {
        if ($$1 == dkv.c && !this.m) {
            return $$0;
        }
        if ($$1 == dkv.b && !this.n) {
            return $$0;
        }
        return this.a($$0, $$1::b);
    }

    private dtc a(dtc $$0, Function<ji, ji> $$1) {
        dtc $$2 = $$0;
        for (ji $$3 : b) {
            if (!this.a($$3)) continue;
            $$2 = (dtc)$$2.a(dky.b($$1.apply($$3)), $$0.c(dky.b($$3)));
        }
        return $$2;
    }

    public static boolean a(dtc $$0, ji $$1) {
        dtt $$2 = dky.b($$1);
        return $$0.b($$2) && $$0.c($$2) != false;
    }

    public static boolean a(dcc $$0, ji $$1, jd $$2, dtc $$3) {
        return dfy.a($$3.l($$0, $$2), $$1.g()) || dfy.a($$3.k($$0, $$2), $$1.g());
    }

    private boolean m() {
        return this.E.d().contains(dts.C);
    }

    private static dtc a(dtc $$0, dtt $$1) {
        dtc $$2 = (dtc)$$0.a($$1, false);
        if (dky.n($$2)) {
            return $$2;
        }
        return dga.a.o();
    }

    public static dtt b(ji $$0) {
        return i.get($$0);
    }

    private static dtc a(dtd<dfy, dtc> $$0) {
        dtc $$1 = $$0.b();
        for (dtt $$2 : i.values()) {
            if (!$$1.b($$2)) continue;
            $$1 = (dtc)$$1.a($$2, false);
        }
        return $$1;
    }

    private static exv o(dtc $$0) {
        exv $$1 = exs.a();
        for (ji $$2 : b) {
            if (!dky.a($$0, $$2)) continue;
            $$1 = exs.a($$1, j.get($$2));
        }
        return $$1.c() ? exs.b() : $$1;
    }

    protected static boolean n(dtc $$0) {
        return Arrays.stream(b).anyMatch($$1 -> dky.a($$0, $$1));
    }

    private static boolean p(dtc $$0) {
        return Arrays.stream(b).anyMatch($$1 -> !dky.a($$0, $$1));
    }

    public abstract dkz c();
}
