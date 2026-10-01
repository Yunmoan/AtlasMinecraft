/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.google.common.annotations.VisibleForTesting
 *  def$b
 *  def$d
 *  def$f
 *  def$g
 *  def$h
 *  dyr
 *  dys
 *  jd
 */
import com.google.common.annotations.VisibleForTesting;
import java.util.List;

/*
 * Exception performing whole class analysis ignored.
 */
public class def {
    private static final boolean b = false;
    private static final float c = 10000.0f;
    @VisibleForTesting
    protected static final int a = 7;

    public static h a(float $$0, float $$1, float $$2, float $$3, float $$4, float $$5) {
        return new h(def.a($$0), def.a($$1), def.a($$2), def.a($$3), def.a($$4), def.a($$5));
    }

    public static d a(float $$0, float $$1, float $$2, float $$3, float $$4, float $$5, float $$6) {
        return new d(def$b.a((float)$$0), def$b.a((float)$$1), def$b.a((float)$$2), def$b.a((float)$$3), def$b.a((float)$$4), def$b.a((float)$$5), def.a($$6));
    }

    public static d a(b $$0, b $$1, b $$2, b $$3, b $$4, b $$5, float $$6) {
        return new d($$0, $$1, $$2, $$3, $$4, $$5, def.a($$6));
    }

    public static long a(float $$0) {
        return (long)($$0 * 10000.0f);
    }

    public static float a(long $$0) {
        return (float)$$0 / 10000.0f;
    }

    public static f a() {
        dyr $$0 = dys.a();
        return new f($$0, $$0, $$0, $$0, $$0, $$0, List.of());
    }

    public static jd a(List<d> $$0, f $$1) {
        return new g($$0, (f)$$1).a.a();
    }
}
