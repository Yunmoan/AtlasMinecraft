/*
 * Decompiled with CFR 0.152.
 * 
 * Could not load the following classes:
 *  com.google.common.annotations.VisibleForTesting
 *  com.google.common.collect.Lists
 *  com.google.common.collect.Maps
 *  com.mojang.datafixers.util.Pair
 *  it.unimi.dsi.fastutil.objects.ObjectArrayList
 *  javax.annotation.Nullable
 */
import com.google.common.annotations.VisibleForTesting;
import com.google.common.collect.Lists;
import com.google.common.collect.Maps;
import com.mojang.datafixers.util.Pair;
import it.unimi.dsi.fastutil.objects.ObjectArrayList;
import java.lang.invoke.MethodHandle;
import java.lang.runtime.ObjectMethods;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.Iterator;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.Optional;
import java.util.stream.Collectors;
import javax.annotation.Nullable;

public class ent {
    public static final String a = "palette";
    public static final String b = "palettes";
    public static final String c = "entities";
    public static final String d = "blocks";
    public static final String e = "pos";
    public static final String f = "state";
    public static final String g = "nbt";
    public static final String h = "pos";
    public static final String i = "blockPos";
    public static final String j = "nbt";
    public static final String k = "size";
    private final List<a> l = Lists.newArrayList();
    private final List<d> m = Lists.newArrayList();
    private kh n = kh.g;
    private String o = "?";

    public kh a() {
        return this.n;
    }

    public void a(String $$0) {
        this.o = $$0;
    }

    public String b() {
        return this.o;
    }

    public void a(dcw $$0, jd $$1, kh $$2, boolean $$3, @Nullable dfy $$4) {
        if ($$2.u() < 1 || $$2.v() < 1 || $$2.w() < 1) {
            return;
        }
        jd $$5 = $$1.a($$2).b(-1, -1, -1);
        ArrayList $$6 = Lists.newArrayList();
        ArrayList $$7 = Lists.newArrayList();
        ArrayList $$8 = Lists.newArrayList();
        jd $$9 = new jd(Math.min($$1.u(), $$5.u()), Math.min($$1.v(), $$5.v()), Math.min($$1.w(), $$5.w()));
        jd $$10 = new jd(Math.max($$1.u(), $$5.u()), Math.max($$1.v(), $$5.v()), Math.max($$1.w(), $$5.w()));
        this.n = $$2;
        for (jd $$11 : jd.c($$9, $$10)) {
            c $$16;
            jd $$12 = $$11.b($$9);
            dtc $$13 = $$0.a_($$11);
            if ($$4 != null && $$13.a($$4)) continue;
            dqh $$14 = $$0.c_($$11);
            if ($$14 != null) {
                c $$15 = new c($$12, $$13, $$14.c($$0.H_()));
            } else {
                $$16 = new c($$12, $$13, null);
            }
            ent.a($$16, $$6, $$7, $$8);
        }
        List<c> $$17 = ent.a($$6, $$7, $$8);
        this.l.clear();
        this.l.add(new a($$17));
        if ($$3) {
            this.a($$0, $$9, $$10);
        } else {
            this.m.clear();
        }
    }

    private static void a(c $$0, List<c> $$1, List<c> $$2, List<c> $$3) {
        if ($$0.c != null) {
            $$2.add($$0);
        } else if (!$$0.b.b().p() && $$0.b.r(dcl.a, jd.c)) {
            $$1.add($$0);
        } else {
            $$3.add($$0);
        }
    }

    private static List<c> a(List<c> $$02, List<c> $$1, List<c> $$2) {
        Comparator<c> $$3 = Comparator.comparingInt($$0 -> $$0.a.v()).thenComparingInt($$0 -> $$0.a.u()).thenComparingInt($$0 -> $$0.a.w());
        $$02.sort($$3);
        $$2.sort($$3);
        $$1.sort($$3);
        ArrayList $$4 = Lists.newArrayList();
        $$4.addAll($$02);
        $$4.addAll($$2);
        $$4.addAll($$1);
        return $$4;
    }

    private void a(dcw $$02, jd $$1, jd $$2) {
        List<bsr> $$3 = $$02.a(bsr.class, ewx.a($$1, $$2), (? super T $$0) -> !($$0 instanceof cmx));
        this.m.clear();
        for (bsr $$4 : $$3) {
            jd $$8;
            exc $$5 = new exc($$4.dt() - (double)$$1.u(), $$4.dv() - (double)$$1.v(), $$4.dz() - (double)$$1.w());
            ub $$6 = new ub();
            $$4.e($$6);
            if ($$4 instanceof cjc) {
                jd $$7 = ((cjc)$$4).t().b($$1);
            } else {
                $$8 = jd.a($$5);
            }
            this.m.add(new d($$5, $$8, $$6.i()));
        }
    }

    public List<c> a(jd $$0, enp $$1, dfy $$2) {
        return this.a($$0, $$1, $$2, true);
    }

    public ObjectArrayList<c> a(jd $$0, enp $$1, dfy $$2, boolean $$3) {
        ObjectArrayList $$4 = new ObjectArrayList();
        ejj $$5 = $$1.g();
        if (this.l.isEmpty()) {
            return $$4;
        }
        for (c $$6 : $$1.a(this.l, $$0).a($$2)) {
            jd $$7;
            jd jd2 = $$7 = $$3 ? ent.a($$1, $$6.a).a($$0) : $$6.a;
            if ($$5 != null && !$$5.b($$7)) continue;
            $$4.add((Object)new c($$7, $$6.b.a($$1.d()), $$6.c));
        }
        return $$4;
    }

    public jd a(enp $$0, jd $$1, enp $$2, jd $$3) {
        jd $$4 = ent.a($$0, $$1);
        jd $$5 = ent.a($$2, $$3);
        return $$4.b($$5);
    }

    public static jd a(enp $$0, jd $$1) {
        return ent.a($$1, $$0.c(), $$0.d(), $$0.e());
    }

    public boolean a(ddl $$0, jd $$1, jd $$2, enp $$3, ayw $$4, int $$5) {
        if (this.l.isEmpty()) {
            return false;
        }
        List<c> $$6 = $$3.a(this.l, $$1).a();
        if ($$6.isEmpty() && ($$3.f() || this.m.isEmpty()) || this.n.u() < 1 || this.n.v() < 1 || this.n.w() < 1) {
            return false;
        }
        ejj $$7 = $$3.g();
        ArrayList $$8 = Lists.newArrayListWithCapacity((int)($$3.j() ? $$6.size() : 0));
        ArrayList $$9 = Lists.newArrayListWithCapacity((int)($$3.j() ? $$6.size() : 0));
        ArrayList $$10 = Lists.newArrayListWithCapacity((int)$$6.size());
        int $$11 = Integer.MAX_VALUE;
        int $$12 = Integer.MAX_VALUE;
        int $$13 = Integer.MAX_VALUE;
        int $$14 = Integer.MIN_VALUE;
        int $$15 = Integer.MIN_VALUE;
        int $$16 = Integer.MIN_VALUE;
        List<c> $$17 = ent.a($$0, $$1, $$2, $$3, $$6);
        for (c $$18 : $$17) {
            dqh $$23;
            jd $$19 = $$18.a;
            if ($$7 != null && !$$7.b($$19)) continue;
            epe $$20 = $$3.j() ? $$0.b_($$19) : null;
            dtc $$21 = $$18.b.a($$3.c()).a($$3.d());
            if ($$18.c != null) {
                dqh $$22 = $$0.c_($$19);
                bqi.a_($$22);
                $$0.a($$19, dga.hW.o(), 20);
            }
            if (!$$0.a($$19, $$21, $$5)) continue;
            $$11 = Math.min($$11, $$19.u());
            $$12 = Math.min($$12, $$19.v());
            $$13 = Math.min($$13, $$19.w());
            $$14 = Math.max($$14, $$19.u());
            $$15 = Math.max($$15, $$19.v());
            $$16 = Math.max($$16, $$19.w());
            $$10.add(Pair.of((Object)$$19, (Object)$$18.c));
            if ($$18.c != null && ($$23 = $$0.c_($$19)) != null) {
                if ($$23 instanceof bqz) {
                    $$18.c.a("LootTableSeed", $$4.g());
                }
                $$23.c($$18.c, $$0.H_());
            }
            if ($$20 == null) continue;
            if ($$21.u().b()) {
                $$9.add($$19);
                continue;
            }
            if (!($$21.b() instanceof dkp)) continue;
            ((dkp)((Object)$$21.b())).a($$0, $$19, $$21, $$20);
            if ($$20.b()) continue;
            $$8.add($$19);
        }
        boolean $$24 = true;
        ji[] $$25 = new ji[]{ji.b, ji.c, ji.f, ji.d, ji.e};
        while ($$24 && !$$8.isEmpty()) {
            $$24 = false;
            Iterator $$26 = $$8.iterator();
            while ($$26.hasNext()) {
                dtc $$32;
                Object $$33;
                jd $$27 = (jd)$$26.next();
                epe $$28 = $$0.b_($$27);
                for (int $$29 = 0; $$29 < $$25.length && !$$28.b(); ++$$29) {
                    jd $$30 = $$27.a($$25[$$29]);
                    epe $$31 = $$0.b_($$30);
                    if (!$$31.b() || $$9.contains($$30)) continue;
                    $$28 = $$31;
                }
                if (!$$28.b() || !(($$33 = ($$32 = $$0.a_($$27)).b()) instanceof dkp)) continue;
                ((dkp)$$33).a($$0, $$27, $$32, $$28);
                $$24 = true;
                $$26.remove();
            }
        }
        if ($$11 <= $$14) {
            if (!$$3.h()) {
                exf $$34 = new exf($$14 - $$11 + 1, $$15 - $$12 + 1, $$16 - $$13 + 1);
                int $$35 = $$11;
                int $$36 = $$12;
                int $$37 = $$13;
                for (Pair $$38 : $$10) {
                    jd $$39 = (jd)$$38.getFirst();
                    ((exl)$$34).c($$39.u() - $$35, $$39.v() - $$36, $$39.w() - $$37);
                }
                ent.a($$0, $$5, $$34, $$35, $$36, $$37);
            }
            for (Pair $$40 : $$10) {
                dqh $$44;
                jd $$41 = (jd)$$40.getFirst();
                if (!$$3.h()) {
                    dtc $$43;
                    dtc $$42 = $$0.a_($$41);
                    if ($$42 != ($$43 = dfy.b($$42, $$0, $$41))) {
                        $$0.a($$41, $$43, $$5 & 0xFFFFFFFE | 0x10);
                    }
                    $$0.b($$41, $$43.b());
                }
                if ($$40.getSecond() == null || ($$44 = $$0.c_($$41)) == null) continue;
                $$44.e();
            }
        }
        if (!$$3.f()) {
            this.a($$0, $$1, $$3.c(), $$3.d(), $$3.e(), $$7, $$3.k());
        }
        return true;
    }

    public static void a(dcx $$0, int $$1, exl $$2, jd $$3) {
        ent.a($$0, $$1, $$2, $$3.u(), $$3.v(), $$3.w());
    }

    public static void a(dcx $$0, int $$1, exl $$2, int $$3, int $$4, int $$5) {
        jd.a $$6 = new jd.a();
        jd.a $$72 = new jd.a();
        $$2.a((ji $$7, int $$8, int $$9, int $$10) -> {
            dtc $$14;
            $$6.d($$3 + $$8, $$4 + $$9, $$5 + $$10);
            $$72.a((kh)$$6, $$7);
            dtc $$11 = $$0.a_($$6);
            dtc $$12 = $$0.a_($$72);
            dtc $$13 = $$11.a($$7, $$12, $$0, (jd)$$6, $$72);
            if ($$11 != $$13) {
                $$0.a((jd)$$6, $$13, $$1 & 0xFFFFFFFE);
            }
            if ($$12 != ($$14 = $$12.a($$7.g(), $$13, $$0, (jd)$$72, $$6))) {
                $$0.a((jd)$$72, $$14, $$1 & 0xFFFFFFFE);
            }
        });
    }

    public static List<c> a(ddl $$0, jd $$1, jd $$2, enp $$3, List<c> $$4) {
        ArrayList<c> $$5 = new ArrayList<c>();
        List<c> $$6 = new ArrayList<c>();
        for (c $$7 : $$4) {
            jd $$8 = ent.a($$3, $$7.a).a($$1);
            c $$9 = new c($$8, $$7.b, $$7.c != null ? $$7.c.i() : null);
            Iterator<enq> $$10 = $$3.i().iterator();
            while ($$9 != null && $$10.hasNext()) {
                $$9 = $$10.next().a((dcz)$$0, $$1, $$2, $$7, $$9, $$3);
            }
            if ($$9 == null) continue;
            $$6.add($$9);
            $$5.add($$7);
        }
        for (enq $$11 : $$3.i()) {
            $$6 = $$11.a($$0, $$1, $$2, $$5, $$6, $$3);
        }
        return $$6;
    }

    private void a(ddl $$0, jd $$1, dkv $$2, dmm $$3, jd $$4, @Nullable ejj $$52, boolean $$6) {
        for (d $$7 : this.m) {
            jd $$8 = ent.a($$7.b, $$2, $$3, $$4).a($$1);
            if ($$52 != null && !$$52.b($$8)) continue;
            ub $$9 = $$7.c.i();
            exc $$10 = ent.a($$7.a, $$2, $$3, $$4);
            exc $$11 = $$10.b($$1.u(), $$1.v(), $$1.w());
            uh $$12 = new uh();
            $$12.add(uc.a($$11.c));
            $$12.add(uc.a($$11.d));
            $$12.add(uc.a($$11.e));
            $$9.a("Pos", $$12);
            $$9.r("UUID");
            ent.a($$0, $$9).ifPresent($$5 -> {
                float $$6 = $$5.a($$3);
                $$5.b($$2.c, $$2.d, $$2.e, $$6 += $$5.a($$2) - $$5.dE(), $$5.dG());
                if ($$6 && $$5 instanceof btp) {
                    ((btp)$$5).a($$0, $$0.d_(jd.a($$11)), btr.d, null);
                }
                $$0.a_((bsr)$$5);
            });
        }
    }

    private static Optional<bsr> a(ddl $$0, ub $$1) {
        try {
            return bsx.a($$1, (dcw)$$0.E());
        }
        catch (Exception $$2) {
            return Optional.empty();
        }
    }

    public kh a(dmm $$0) {
        switch ($$0) {
            case d: 
            case b: {
                return new kh(this.n.w(), this.n.v(), this.n.u());
            }
        }
        return this.n;
    }

    public static jd a(jd $$0, dkv $$1, dmm $$2, jd $$3) {
        int $$4 = $$0.u();
        int $$5 = $$0.v();
        int $$6 = $$0.w();
        boolean $$7 = true;
        switch ($$1) {
            case b: {
                $$6 = -$$6;
                break;
            }
            case c: {
                $$4 = -$$4;
                break;
            }
            default: {
                $$7 = false;
            }
        }
        int $$8 = $$3.u();
        int $$9 = $$3.w();
        switch ($$2) {
            case c: {
                return new jd($$8 + $$8 - $$4, $$5, $$9 + $$9 - $$6);
            }
            case d: {
                return new jd($$8 - $$9 + $$6, $$5, $$8 + $$9 - $$4);
            }
            case b: {
                return new jd($$8 + $$9 - $$6, $$5, $$9 - $$8 + $$4);
            }
        }
        return $$7 ? new jd($$4, $$5, $$6) : $$0;
    }

    public static exc a(exc $$0, dkv $$1, dmm $$2, jd $$3) {
        double $$4 = $$0.c;
        double $$5 = $$0.d;
        double $$6 = $$0.e;
        boolean $$7 = true;
        switch ($$1) {
            case b: {
                $$6 = 1.0 - $$6;
                break;
            }
            case c: {
                $$4 = 1.0 - $$4;
                break;
            }
            default: {
                $$7 = false;
            }
        }
        int $$8 = $$3.u();
        int $$9 = $$3.w();
        switch ($$2) {
            case c: {
                return new exc((double)($$8 + $$8 + 1) - $$4, $$5, (double)($$9 + $$9 + 1) - $$6);
            }
            case d: {
                return new exc((double)($$8 - $$9) + $$6, $$5, (double)($$8 + $$9 + 1) - $$4);
            }
            case b: {
                return new exc((double)($$8 + $$9 + 1) - $$6, $$5, (double)($$9 - $$8) + $$4);
            }
        }
        return $$7 ? new exc($$4, $$5, $$6) : $$0;
    }

    public jd a(jd $$0, dkv $$1, dmm $$2) {
        return ent.a($$0, $$1, $$2, this.a().u(), this.a().w());
    }

    public static jd a(jd $$0, dkv $$1, dmm $$2, int $$3, int $$4) {
        int $$5 = $$1 == dkv.c ? --$$3 : 0;
        int $$6 = $$1 == dkv.b ? --$$4 : 0;
        jd $$7 = $$0;
        switch ($$2) {
            case a: {
                $$7 = $$0.b($$5, 0, $$6);
                break;
            }
            case b: {
                $$7 = $$0.b($$4 - $$6, 0, $$5);
                break;
            }
            case c: {
                $$7 = $$0.b($$3 - $$5, 0, $$4 - $$6);
                break;
            }
            case d: {
                $$7 = $$0.b($$6, 0, $$3 - $$5);
            }
        }
        return $$7;
    }

    public ejj b(enp $$0, jd $$1) {
        return this.a($$1, $$0.d(), $$0.e(), $$0.c());
    }

    public ejj a(jd $$0, dmm $$1, jd $$2, dkv $$3) {
        return ent.a($$0, $$1, $$2, $$3, this.n);
    }

    @VisibleForTesting
    protected static ejj a(jd $$0, dmm $$1, jd $$2, dkv $$3, kh $$4) {
        kh $$5 = $$4.c(-1, -1, -1);
        jd $$6 = ent.a(jd.c, $$3, $$1, $$2);
        jd $$7 = ent.a(jd.c.a($$5), $$3, $$1, $$2);
        return ejj.a($$6, $$7).a((kh)$$0);
    }

    public ub a(ub $$0) {
        if (this.l.isEmpty()) {
            $$0.a(d, new uh());
            $$0.a(a, new uh());
        } else {
            ArrayList $$1 = Lists.newArrayList();
            b $$2 = new b();
            $$1.add($$2);
            for (int $$3 = 1; $$3 < this.l.size(); ++$$3) {
                $$1.add(new b());
            }
            uh $$4 = new uh();
            List<c> $$5 = this.l.get(0).a();
            for (int $$6 = 0; $$6 < $$5.size(); ++$$6) {
                c $$7 = $$5.get($$6);
                ub $$8 = new ub();
                $$8.a("pos", this.a($$7.a.u(), $$7.a.v(), $$7.a.w()));
                int $$9 = $$2.a($$7.b);
                $$8.a(f, $$9);
                if ($$7.c != null) {
                    $$8.a("nbt", $$7.c);
                }
                $$4.add($$8);
                for (int $$10 = 1; $$10 < this.l.size(); ++$$10) {
                    b $$11 = (b)$$1.get($$10);
                    $$11.a(this.l.get((int)$$10).a().get((int)$$6).b, $$9);
                }
            }
            $$0.a(d, $$4);
            if ($$1.size() == 1) {
                uh $$12 = new uh();
                for (dtc $$13 : $$2) {
                    $$12.add(uq.a($$13));
                }
                $$0.a(a, $$12);
            } else {
                uh $$14 = new uh();
                for (b $$15 : $$1) {
                    uh $$16 = new uh();
                    for (dtc $$17 : $$15) {
                        $$16.add(uq.a($$17));
                    }
                    $$14.add($$16);
                }
                $$0.a(b, $$14);
            }
        }
        uh $$18 = new uh();
        for (d $$19 : this.m) {
            ub $$20 = new ub();
            $$20.a("pos", this.a($$19.a.c, $$19.a.d, $$19.a.e));
            $$20.a(i, this.a($$19.b.u(), $$19.b.v(), $$19.b.w()));
            if ($$19.c != null) {
                $$20.a("nbt", $$19.c);
            }
            $$18.add($$20);
        }
        $$0.a(c, $$18);
        $$0.a(k, this.a(this.n.u(), this.n.v(), this.n.w()));
        return uq.e($$0);
    }

    public void a(jn<dfy> $$0, ub $$1) {
        this.l.clear();
        this.m.clear();
        uh $$2 = $$1.c(k, 3);
        this.n = new kh($$2.e(0), $$2.e(1), $$2.e(2));
        uh $$3 = $$1.c(d, 10);
        if ($$1.b(b, 9)) {
            uh $$4 = $$1.c(b, 9);
            for (int $$5 = 0; $$5 < $$4.size(); ++$$5) {
                this.a($$0, $$4.b($$5), $$3);
            }
        } else {
            this.a($$0, $$1.c(a, 10), $$3);
        }
        uh $$6 = $$1.c(c, 10);
        for (int $$7 = 0; $$7 < $$6.size(); ++$$7) {
            ub $$8 = $$6.a($$7);
            uh $$9 = $$8.c("pos", 6);
            exc $$10 = new exc($$9.h(0), $$9.h(1), $$9.h(2));
            uh $$11 = $$8.c(i, 3);
            jd $$12 = new jd($$11.e(0), $$11.e(1), $$11.e(2));
            if (!$$8.e("nbt")) continue;
            ub $$13 = $$8.p("nbt");
            this.m.add(new d($$10, $$12, $$13));
        }
    }

    private void a(jn<dfy> $$0, uh $$1, uh $$2) {
        b $$3 = new b();
        for (int $$4 = 0; $$4 < $$1.size(); ++$$4) {
            $$3.a(uq.a($$0, $$1.a($$4)), $$4);
        }
        ArrayList $$5 = Lists.newArrayList();
        ArrayList $$6 = Lists.newArrayList();
        ArrayList $$7 = Lists.newArrayList();
        for (int $$8 = 0; $$8 < $$2.size(); ++$$8) {
            ub $$14;
            ub $$9 = $$2.a($$8);
            uh $$10 = $$9.c("pos", 3);
            jd $$11 = new jd($$10.e(0), $$10.e(1), $$10.e(2));
            dtc $$12 = $$3.a($$9.h(f));
            if ($$9.e("nbt")) {
                ub $$13 = $$9.p("nbt");
            } else {
                $$14 = null;
            }
            c $$15 = new c($$11, $$12, $$14);
            ent.a($$15, $$5, $$6, $$7);
        }
        List<c> $$16 = ent.a($$5, $$6, $$7);
        this.l.add(new a($$16));
    }

    private uh a(int ... $$0) {
        uh $$1 = new uh();
        for (int $$2 : $$0) {
            $$1.add(ug.a($$2));
        }
        return $$1;
    }

    private uh a(double ... $$0) {
        uh $$1 = new uh();
        for (double $$2 : $$0) {
            $$1.add(uc.a($$2));
        }
        return $$1;
    }

    public static final class c
    extends Record {
        final jd a;
        final dtc b;
        @Nullable
        final ub c;

        public c(jd $$0, dtc $$1, @Nullable ub $$2) {
            this.a = $$0;
            this.b = $$1;
            this.c = $$2;
        }

        @Override
        public String toString() {
            return String.format(Locale.ROOT, "<StructureBlockInfo | %s | %s | %s>", this.a, this.b, this.c);
        }

        @Override
        public final int hashCode() {
            return (int)ObjectMethods.bootstrap("hashCode", new MethodHandle[]{c.class, "pos;state;nbt", "a", "b", "c"}, this);
        }

        @Override
        public final boolean equals(Object $$0) {
            return (boolean)ObjectMethods.bootstrap("equals", new MethodHandle[]{c.class, "pos;state;nbt", "a", "b", "c"}, this, $$0);
        }

        public jd a() {
            return this.a;
        }

        public dtc b() {
            return this.b;
        }

        @Nullable
        public ub c() {
            return this.c;
        }
    }

    public static final class a {
        private final List<c> a;
        private final Map<dfy, List<c>> b = Maps.newHashMap();

        a(List<c> $$0) {
            this.a = $$0;
        }

        public List<c> a() {
            return this.a;
        }

        public List<c> a(dfy $$02) {
            return this.b.computeIfAbsent($$02, $$0 -> this.a.stream().filter($$1 -> $$1.b.a((dfy)$$0)).collect(Collectors.toList()));
        }
    }

    public static class d {
        public final exc a;
        public final jd b;
        public final ub c;

        public d(exc $$0, jd $$1, ub $$2) {
            this.a = $$0;
            this.b = $$1;
            this.c = $$2;
        }
    }

    static class b
    implements Iterable<dtc> {
        public static final dtc a = dga.a.o();
        private final js<dtc> b = new js(16);
        private int c;

        b() {
        }

        public int a(dtc $$0) {
            int $$1 = this.b.a($$0);
            if ($$1 == -1) {
                $$1 = this.c++;
                this.b.a($$0, $$1);
            }
            return $$1;
        }

        @Nullable
        public dtc a(int $$0) {
            dtc $$1 = this.b.a($$0);
            return $$1 == null ? a : $$1;
        }

        @Override
        public Iterator<dtc> iterator() {
            return this.b.iterator();
        }

        public void a(dtc $$0, int $$1) {
            this.b.a($$0, $$1);
        }
    }
}
