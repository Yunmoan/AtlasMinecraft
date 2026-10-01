/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.serialization.MapCodec
 */
import com.mojang.serialization.MapCodec;

public class dkw
extends dfy
implements dgb {
    public static final MapCodec<dkw> a = dkw.b(dkw::new);

    public MapCodec<dkw> a() {
        return a;
    }

    public dkw(dtb.d $$0) {
        super($$0);
    }

    @Override
    public boolean b(dcz $$0, jd $$1, dtc $$2) {
        return $$0.a_($$1.d()).i();
    }

    @Override
    public boolean a(dcw $$0, ayw $$1, jd $$2, dtc $$3) {
        return true;
    }

    @Override
    public void a(aqu $$02, ayw $$1, jd $$2, dtc $$32) {
        $$02.H_().c(lu.aI).flatMap($$0 -> $$0.b(rp.n)).ifPresent($$3 -> ((ebq)$$3.a()).a($$02, $$02.l().g(), $$1, $$2.d()));
    }

    @Override
    public dgb.a au_() {
        return dgb.a.a;
    }
}
