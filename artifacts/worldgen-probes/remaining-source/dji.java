/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.serialization.MapCodec
 */
import com.mojang.serialization.MapCodec;
import java.util.function.ToIntFunction;

public class dji
extends dky
implements dgb,
dna {
    public static final MapCodec<dji> a = dji.b(dji::new);
    private static final dtt c = dts.C;
    private final dkz d = new dkz(this);

    public MapCodec<dji> a() {
        return a;
    }

    public dji(dtb.d $$0) {
        super($$0);
        this.k((dtc)this.o().a(c, false));
    }

    public static ToIntFunction<dtc> b(int $$0) {
        return $$1 -> dky.n($$1) ? $$0 : 0;
    }

    @Override
    protected void a(dtd.a<dfy, dtc> $$0) {
        super.a($$0);
        $$0.a(new duf[]{c});
    }

    @Override
    protected dtc a(dtc $$0, ji $$1, dtc $$2, dcx $$3, jd $$4, jd $$5) {
        if ($$0.c(c).booleanValue()) {
            $$3.a($$4, epf.c, epf.c.a($$3));
        }
        return super.a($$0, $$1, $$2, $$3, $$4, $$5);
    }

    @Override
    protected boolean a(dtc $$0, cyd $$1) {
        return !$$1.n().a(cut.fW) || super.a($$0, $$1);
    }

    @Override
    public boolean b(dcz $$0, jd $$1, dtc $$2) {
        return ji.a().anyMatch($$3 -> this.d.a($$2, $$0, $$1, $$3.g()));
    }

    @Override
    public boolean a(dcw $$0, ayw $$1, jd $$2, dtc $$3) {
        return true;
    }

    @Override
    public void a(aqu $$0, ayw $$1, jd $$2, dtc $$3) {
        this.d.a($$3, (dcx)$$0, $$2, $$1);
    }

    @Override
    protected epe b_(dtc $$0) {
        if ($$0.c(c).booleanValue()) {
            return epf.c.a(false);
        }
        return super.b_($$0);
    }

    @Override
    protected boolean a_(dtc $$0, dcc $$1, jd $$2) {
        return $$0.u().c();
    }

    @Override
    public dkz c() {
        return this.d;
    }
}
