/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  cmx
 *  com.mojang.serialization.MapCodec
 *  dcc
 *  dcx
 *  dga
 *  djm
 *  djn
 *  dkp
 *  dtb$d
 *  dtc
 *  epd
 *  epe
 *  epf
 *  exs
 *  javax.annotation.Nullable
 *  jd
 *  ji
 */
import com.mojang.serialization.MapCodec;
import javax.annotation.Nullable;

public class dkd
extends djm
implements dkp {
    public static final MapCodec<dkd> c = dkd.b(dkd::new);

    public MapCodec<dkd> a() {
        return c;
    }

    protected dkd(dtb.d $$0) {
        super($$0, ji.b, exs.b(), true);
    }

    protected djn c() {
        return (djn)dga.mc;
    }

    protected epe b_(dtc $$0) {
        return epf.c.a(false);
    }

    protected boolean m(dtc $$0) {
        return this.c().m($$0);
    }

    public boolean a(@Nullable cmx $$0, dcc $$1, jd $$2, dtc $$3, epd $$4) {
        return false;
    }

    public boolean a(dcx $$0, jd $$1, dtc $$2, epe $$3) {
        return false;
    }
}
