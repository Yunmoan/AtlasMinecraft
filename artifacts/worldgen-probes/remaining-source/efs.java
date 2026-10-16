/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.mojang.datafixers.kinds.Applicative
 *  com.mojang.serialization.MapCodec
 *  com.mojang.serialization.codecs.RecordCodecBuilder
 */
import com.mojang.datafixers.kinds.Applicative;
import com.mojang.serialization.MapCodec;
import com.mojang.serialization.codecs.RecordCodecBuilder;

public class efs
extends efu {
    public static final MapCodec<efs> a = RecordCodecBuilder.mapCodec($$0 -> efs.b($$0).apply((Applicative)$$0, efs::new));

    public efs(bpw $$0, bpw $$1) {
        super($$0, $$1);
    }

    @Override
    protected efv<?> a() {
        return efv.i;
    }

    @Override
    protected void a(ddc $$0, efu.b $$1, ayw $$2, efe $$3, int $$4, efu.a $$5, int $$6, int $$7, int $$8) {
        jd $$9 = $$5.a().b($$8);
        boolean $$10 = $$5.c();
        if ($$10) {
            this.a($$0, $$1, $$2, $$3, $$9, $$7 + 2, -1, $$10);
            this.a($$0, $$1, $$2, $$3, $$9, $$7 + 3, 0, $$10);
            this.a($$0, $$1, $$2, $$3, $$9, $$7 + 2, 1, $$10);
            if ($$2.h()) {
                this.a($$0, $$1, $$2, $$3, $$9, $$7, 2, $$10);
            }
        } else {
            this.a($$0, $$1, $$2, $$3, $$9, $$7 + 2, -1, $$10);
            this.a($$0, $$1, $$2, $$3, $$9, $$7 + 1, 0, $$10);
        }
    }

    @Override
    public int a(ayw $$0, int $$1, efe $$2) {
        return 4;
    }

    @Override
    protected boolean b(ayw $$0, int $$1, int $$2, int $$3, int $$4, boolean $$5) {
        if (!($$2 != 0 || !$$5 || $$1 != -$$4 && $$1 < $$4 || $$3 != -$$4 && $$3 < $$4)) {
            return true;
        }
        return super.b($$0, $$1, $$2, $$3, $$4, $$5);
    }

    @Override
    protected boolean a(ayw $$0, int $$1, int $$2, int $$3, int $$4, boolean $$5) {
        if ($$2 == -1 && !$$5) {
            return $$1 == $$4 && $$3 == $$4;
        }
        if ($$2 == 1) {
            return $$1 + $$3 > $$4 * 2 - 2;
        }
        return false;
    }
}
