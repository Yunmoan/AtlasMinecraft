/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  awk
 *  ayw
 *  cmx
 *  com.mojang.serialization.MapCodec
 *  cyd
 *  dcc
 *  dcx
 *  dfy
 *  dga
 *  djn
 *  dkp
 *  dtb$d
 *  dtc
 *  epd
 *  epe
 *  epf
 *  exv
 *  javax.annotation.Nullable
 *  jd
 *  ji
 */
import com.mojang.serialization.MapCodec;
import javax.annotation.Nullable;

public class dkc
extends djn
implements dkp {
    public static final MapCodec<dkc> c = dkc.b(dkc::new);
    protected static final exv g = dfy.a((double)0.0, (double)0.0, (double)0.0, (double)16.0, (double)9.0, (double)16.0);
    private static final double h = 0.14;

    public MapCodec<dkc> a() {
        return c;
    }

    protected dkc(dtb.d $$0) {
        super($$0, ji.b, g, true, 0.14);
    }

    protected boolean g(dtc $$0) {
        return $$0.a(dga.G);
    }

    protected dfy b() {
        return dga.md;
    }

    protected boolean m(dtc $$0) {
        return !$$0.a(dga.kJ);
    }

    public boolean a(@Nullable cmx $$0, dcc $$1, jd $$2, dtc $$3, epd $$4) {
        return false;
    }

    public boolean a(dcx $$0, jd $$1, dtc $$2, epe $$3) {
        return false;
    }

    protected int a(ayw $$0) {
        return 1;
    }

    @Nullable
    public dtc a(cyd $$0) {
        epe $$1 = $$0.q().b_($$0.a());
        if ($$1.a(awk.a) && $$1.e() == 8) {
            return super.a($$0);
        }
        return null;
    }

    protected epe b_(dtc $$0) {
        return epf.c.a(false);
    }
}
