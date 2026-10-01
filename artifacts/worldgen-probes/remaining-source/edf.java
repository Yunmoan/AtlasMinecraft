/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.serialization.Codec
 */
import com.mojang.serialization.Codec;

public class edf
extends ece<eew> {
    public edf(Codec<eew> $$0) {
        super($$0);
    }

    @Override
    public boolean a(ecg<eew> $$0) {
        dds $$1 = $$0.b();
        jd $$2 = $$0.e();
        eew $$3 = $$0.f();
        for (eeq.a $$4 : $$3.b) {
            if (!$$4.b.a($$1.a_($$2), $$0.d())) continue;
            $$1.a($$2, $$4.c, 2);
            break;
        }
        return true;
    }
}
