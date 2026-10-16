package atlasprobe;

import java.io.ByteArrayOutputStream;
import java.io.DataOutputStream;
import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.util.Base64;

/** Saved constructor defaults for the village Jigsaw campfire replacement. */
public final class VillageCampfireBlockEntityProbe {
    static Class<?> cls(String name) throws Exception { return Class.forName(name); }
    public static void main(String[] args) throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);
        cls("ab").getMethod("a",cls("ae")).invoke(null,version);
        cls("akt").getMethod("a").invoke(null);
        Object registries=IglooProbeSupport.registryAccess();
        Object position=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(0,0,0);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        for(String name:new String[]{"campfire","soul_campfire"}) {
            Object state=RuinedPortalProbeSupport.block(name);
            Object block=cls("dtb$a").getMethod("b").invoke(state);
            Object entity=cls("diq").getMethod("a",cls("jd"),cls("dtc")).invoke(block,position,state);
            Object saved=cls("dqh").getMethod("b",cls("jo$a")).invoke(entity,registries);
            for(String field:new String[]{"x","y","z"}) cls("ub").getMethod("r",String.class).invoke(saved,field);
            ByteArrayOutputStream bytes=new ByteArrayOutputStream();
            cls("uo").getMethod("a",cls("uy"),java.io.DataOutput.class).invoke(null,saved,new DataOutputStream(bytes));
            System.out.printf("campfire_default minecraft:%s %s%n",name,Base64.getEncoder().encodeToString(bytes.toByteArray()));
        }
    }
}
