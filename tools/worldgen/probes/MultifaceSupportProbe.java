package atlasprobe;

import java.io.*;
import java.lang.reflect.*;

/** Exact support-or-collision face masks used by MultifaceBlock.canAttachTo. */
public final class MultifaceSupportProbe {
    public static void main(String[] args) throws Exception {
        Object version=Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a",Class.forName("ae")).invoke(null,version);
        Class.forName("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Object world=Class.forName("dcl").getField("a").get(null);
        Object origin=Class.forName("jd").getConstructor(int.class,int.class,int.class).newInstance(0,0,0);
        Method decode=Class.forName("dfy").getMethod("a",int.class);
        Method support=Class.forName("dtb$a").getMethod("l",Class.forName("dcc"),Class.forName("jd"));
        Method collision=Class.forName("dtb$a").getMethod("k",Class.forName("dcc"),Class.forName("jd"));
        Method full=Class.forName("dfy").getMethod("a",Class.forName("exv"),Class.forName("ji"));
        Object[] faces=(Object[])Class.forName("ji").getMethod("values").invoke(null);
        for(int id=0;id<26684;id++) {
            Object state=decode.invoke(null,id); int mask=0;
            Object s=support.invoke(state,world,origin),c=collision.invoke(state,world,origin);
            for(int i=0;i<6;i++) if((boolean)full.invoke(null,s,faces[i])||(boolean)full.invoke(null,c,faces[i])) mask|=1<<i;
            System.out.println(mask);
        }
    }
}
