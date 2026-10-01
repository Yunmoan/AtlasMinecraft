/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.google.common.annotations.VisibleForTesting
 *  javax.annotation.Nullable
 */
import com.google.common.annotations.VisibleForTesting;
import java.lang.invoke.MethodHandle;
import java.lang.runtime.ObjectMethods;
import java.util.Optional;
import javax.annotation.Nullable;

public class dkz {
    public static final e[] a = new e[]{e.a, e.b, e.c};
    private final b b;

    public dkz(dky $$0) {
        this(new a($$0));
    }

    public dkz(b $$0) {
        this.b = $$0;
    }

    public boolean a(dtc $$0, dcc $$1, jd $$2, ji $$3) {
        return ji.a().anyMatch($$4 -> this.a($$0, $$1, $$2, $$3, (ji)$$4, this.b::a).isPresent());
    }

    public Optional<c> a(dtc $$0, dcx $$12, jd $$2, ayw $$3) {
        return ji.a($$3).stream().filter($$1 -> this.b.b($$0, (ji)$$1)).map($$4 -> this.a($$0, $$12, $$2, (ji)$$4, $$3, false)).filter(Optional::isPresent).findFirst().orElse(Optional.empty());
    }

    public long a(dtc $$0, dcx $$12, jd $$2, boolean $$3) {
        return ji.a().filter($$1 -> this.b.b($$0, (ji)$$1)).map($$4 -> this.a($$0, $$12, $$2, (ji)$$4, $$3)).reduce(0L, Long::sum);
    }

    public Optional<c> a(dtc $$0, dcx $$1, jd $$2, ji $$3, ayw $$4, boolean $$52) {
        return ji.a($$4).stream().map($$5 -> this.a($$0, $$1, $$2, $$3, (ji)$$5, $$52)).filter(Optional::isPresent).findFirst().orElse(Optional.empty());
    }

    private long a(dtc $$0, dcx $$1, jd $$2, ji $$3, boolean $$4) {
        return ji.a().map($$5 -> this.a($$0, $$1, $$2, $$3, (ji)$$5, $$4)).filter(Optional::isPresent).count();
    }

    @VisibleForTesting
    public Optional<c> a(dtc $$0, dcx $$1, jd $$22, ji $$3, ji $$4, boolean $$5) {
        return this.a($$0, (dcc)$$1, $$22, $$3, $$4, this.b::a).flatMap($$2 -> this.a($$1, (c)$$2, $$5));
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

    public static class a
    implements b {
        protected dky a;

        public a(dky $$0) {
            this.a = $$0;
        }

        @Override
        @Nullable
        public dtc a(dtc $$0, dcc $$1, jd $$2, ji $$3) {
            return this.a.c($$0, $$1, $$2, $$3);
        }

        protected boolean a(dcc $$0, jd $$1, jd $$2, ji $$3, dtc $$4) {
            return $$4.i() || $$4.a(this.a) || $$4.a(dga.G) && $$4.u().b();
        }

        @Override
        public boolean a(dcc $$0, jd $$1, c $$2) {
            dtc $$3 = $$0.a_($$2.a());
            return this.a($$0, $$1, $$2.a(), $$2.b(), $$3) && this.a.a($$0, $$3, $$2.a(), $$2.b());
        }
    }

    public static interface b {
        @Nullable
        public dtc a(dtc var1, dcc var2, jd var3, ji var4);

        public boolean a(dcc var1, jd var2, c var3);

        default public e[] a() {
            return a;
        }

        default public boolean a(dtc $$0, ji $$1) {
            return dky.a($$0, $$1);
        }

        default public boolean a(dtc $$0) {
            return false;
        }

        default public boolean b(dtc $$0, ji $$1) {
            return this.a($$0) || this.a($$0, $$1);
        }

        default public boolean a(dcx $$0, c $$1, dtc $$2, boolean $$3) {
            dtc $$4 = this.a($$2, $$0, $$1.a(), $$1.b());
            if ($$4 != null) {
                if ($$3) {
                    $$0.y($$1.a()).e($$1.a());
                }
                return $$0.a($$1.a(), $$4, 2);
            }
            return false;
        }
    }

    @FunctionalInterface
    public static interface d {
        public boolean test(dcc var1, jd var2, c var3);
    }

    public static abstract sealed class e
    extends Enum<e> {
        public static final /* enum */ e a = new e(){

            @Override
            public c a(jd $$0, ji $$1, ji $$2) {
                return new c($$0, $$1);
            }
        };
        public static final /* enum */ e b = new e(){

            @Override
            public c a(jd $$0, ji $$1, ji $$2) {
                return new c($$0.a($$1), $$2);
            }
        };
        public static final /* enum */ e c = new e(){

            @Override
            public c a(jd $$0, ji $$1, ji $$2) {
                return new c($$0.a($$1).a($$2), $$1.g());
            }
        };
        private static final /* synthetic */ e[] d;

        public static e[] values() {
            return (e[])d.clone();
        }

        public static e valueOf(String $$0) {
            return Enum.valueOf(e.class, $$0);
        }

        public abstract c a(jd var1, ji var2, ji var3);

        private static /* synthetic */ e[] a() {
            return new e[]{a, b, c};
        }

        static {
            d = e.a();
        }
    }

    public record c(jd a, ji b) {
        @Override
        public final String toString() {
            return ObjectMethods.bootstrap("toString", new MethodHandle[]{c.class, "pos;face", "a", "b"}, this);
        }

        @Override
        public final int hashCode() {
            return (int)ObjectMethods.bootstrap("hashCode", new MethodHandle[]{c.class, "pos;face", "a", "b"}, this);
        }

        @Override
        public final boolean equals(Object $$0) {
            return (boolean)ObjectMethods.bootstrap("equals", new MethodHandle[]{c.class, "pos;face", "a", "b"}, this, $$0);
        }
    }
}
