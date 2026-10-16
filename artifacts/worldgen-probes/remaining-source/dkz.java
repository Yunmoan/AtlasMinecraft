/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  ayw
 *  com.google.common.annotations.VisibleForTesting
 *  dcc
 *  dcx
 *  dkz$a
 *  dkz$b
 *  dkz$c
 *  dkz$d
 *  dkz$e
 *  dtc
 *  jd
 *  ji
 */
import com.google.common.annotations.VisibleForTesting;
import java.util.Optional;

public class dkz {
    public static final e[] a = new e[]{e.a, e.b, e.c};
    private final b b;

    public dkz(dky $$0) {
        this((b)new a($$0));
    }

    public dkz(b $$0) {
        this.b = $$0;
    }

    public boolean a(dtc $$0, dcc $$1, jd $$2, ji $$3) {
        return ji.a().anyMatch($$4 -> this.a($$0, $$1, $$2, $$3, (ji)$$4, (arg_0, arg_1, arg_2) -> ((b)this.b).a(arg_0, arg_1, arg_2)).isPresent());
    }

    public Optional<c> a(dtc $$0, dcx $$12, jd $$2, ayw $$3) {
        return ji.a((ayw)$$3).stream().filter($$1 -> this.b.b($$0, $$1)).map($$4 -> this.a($$0, $$12, $$2, (ji)$$4, $$3, false)).filter(Optional::isPresent).findFirst().orElse(Optional.empty());
    }

    public long a(dtc $$0, dcx $$12, jd $$2, boolean $$3) {
        return ji.a().filter($$1 -> this.b.b($$0, $$1)).map($$4 -> this.a($$0, $$12, $$2, (ji)$$4, $$3)).reduce(0L, Long::sum);
    }

    public Optional<c> a(dtc $$0, dcx $$1, jd $$2, ji $$3, ayw $$4, boolean $$52) {
        return ji.a((ayw)$$4).stream().map($$5 -> this.a($$0, $$1, $$2, $$3, (ji)$$5, $$52)).filter(Optional::isPresent).findFirst().orElse(Optional.empty());
    }

    private long a(dtc $$0, dcx $$1, jd $$2, ji $$3, boolean $$4) {
        return ji.a().map($$5 -> this.a($$0, $$1, $$2, $$3, (ji)$$5, $$4)).filter(Optional::isPresent).count();
    }

    @VisibleForTesting
    public Optional<c> a(dtc $$0, dcx $$1, jd $$22, ji $$3, ji $$4, boolean $$5) {
        return this.a($$0, (dcc)$$1, $$22, $$3, $$4, (arg_0, arg_1, arg_2) -> ((b)this.b).a(arg_0, arg_1, arg_2)).flatMap($$2 -> this.a($$1, (c)$$2, $$5));
    }

    public Optional<c> a(dtc $$0, dcc $$1, jd $$2, ji $$3, ji $$4, d $$5) {
        if ($$4.o() == $$3.o()) {
            return Optional.empty();
        }
        if (!(this.b.a($$0) || this.b.a($$0, $$3) && !this.b.a($$0, $$4))) {
            return Optional.empty();
        }
        for (e $$6 : this.b.a()) {
            c $$7 = $$6.a($$2, $$4, $$3);
            if (!$$5.test($$1, $$2, $$7)) continue;
            return Optional.of($$7);
        }
        return Optional.empty();
    }

    public Optional<c> a(dcx $$0, c $$1, boolean $$2) {
        dtc $$3 = $$0.a_($$1.a());
        if (this.b.a($$0, $$1, $$3, $$2)) {
            return Optional.of($$1);
        }
        return Optional.empty();
    }
}
