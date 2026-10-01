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

public class dmn
extends dgh
implements dgb {
    public static final MapCodec<dmn> e = RecordCodecBuilder.mapCodec($$02 -> $$02.group((App)dsr.a.fieldOf("tree").forGetter($$0 -> $$0.i), dmn.u()).apply((Applicative)$$02, dmn::new));
    public static final duc f = dts.aU;
    protected static final float g = 6.0f;
    protected static final exv h = dfy.a(2.0, 0.0, 2.0, 14.0, 12.0, 14.0);
    protected final dsr i;

    public MapCodec<? extends dmn> a() {
        return e;
    }

    protected dmn(dsr $$0, dtb.d $$1) {
        super($$1);
        this.i = $$0;
        this.k((dtc)((dtc)this.E.b()).a(f, 0));
    }

    @Override
    protected exv a(dtc $$0, dcc $$1, jd $$2, exh $$3) {
        return h;
    }

    @Override
    protected void b(dtc $$0, aqu $$1, jd $$2, ayw $$3) {
        if ($$1.A($$2.d()) >= 9 && $$3.a(7) == 0) {
            this.a($$1, $$2, $$0, $$3);
        }
    }

    public void a(aqu $$0, jd $$1, dtc $$2, ayw $$3) {
        if ($$2.c(f) == 0) {
            $$0.a($$1, (dtc)$$2.a(f), 4);
        } else {
            this.i.a($$0, $$0.l().g(), $$1, $$2, $$3);
        }
    }

    @Override
    public boolean b(dcz $$0, jd $$1, dtc $$2) {
        return true;
    }

    @Override
    public boolean a(dcw $$0, ayw $$1, jd $$2, dtc $$3) {
        return (double)$$0.z.i() < 0.45;
    }

    @Override
    public void a(aqu $$0, ayw $$1, jd $$2, dtc $$3) {
        this.a($$0, $$2, $$3, $$1);
    }

    @Override
    protected void a(dtd.a<dfy, dtc> $$0) {
        $$0.a(new duf[]{f});
    }
}
