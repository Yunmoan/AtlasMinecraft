package atlasprobe;
import java.io.*;
import java.lang.reflect.*;
import java.util.Base64;
public final class PortalRulesProbe {
    public static void main(String[] args) throws Exception {
        Object version=Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a",Class.forName("ae")).invoke(null,version);
        Class.forName("akt").getMethod("a").invoke(null);
        Method decode=Class.forName("dfy").getMethod("a",int.class),replace=Class.forName("dtb$a").getMethod("r"),solid=Class.forName("dtb$a").getMethod("e"),fluid=Class.forName("dtb$a").getMethod("u"),empty=Class.forName("epe").getMethod("c");
        byte[] flags=new byte[(26684+3)/4];
        for(int id=0;id<26684;id++) {
            Object state=decode.invoke(null,id);
            int bits=((boolean)replace.invoke(state)&&(boolean)empty.invoke(fluid.invoke(state))?1:0)|((boolean)solid.invoke(state)?2:0);
            flags[id/4]|=(byte)(bits<<((id%4)*2));
        }
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        System.out.println("portal_block_rules 26684 "+Base64.getEncoder().encodeToString(flags));
    }
}
