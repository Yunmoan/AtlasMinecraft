package atlasprobe;

import com.google.gson.Gson;
import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

/** Exhaustive original classifiers used by ruined portal processors/shapes. */
public final class RuinedPortalClassifiersProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        RuinedPortalProbeSupport.bindTags();
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        Object world=cls("dcl").getField("a").get(null),origin=cls("jd").getField("c").get(null),settings=cls("enp").getConstructor().newInstance();
        Object[] directions=cls("ji").getEnumConstants();
        var wallCoverage=cls("dox").getDeclaredMethod("a",cls("exv"),cls("exv"));wallCoverage.setAccessible(true);
        List<int[]> states=new ArrayList<>();
        for(int id=0;id<26684;id++) {
            Object state=cls("dfy").getMethod("a",int.class).invoke(null,id);
            int flags=0;
            String[] tags={"bP","K","L","M","aP"};
            for(int bit=0;bit<tags.length;bit++)if((Boolean)cls("dtb$a").getMethod("a",cls("awu")).invoke(state,cls("awe").getField(tags[bit]).get(null)))flags|=1<<bit;
            Object outline=cls("dtb$a").getMethod("j",cls("dcc"),cls("jd")).invoke(state,world,origin);
            if((Boolean)cls("dfy").getMethod("a",cls("exv")).invoke(null,outline))flags|=32;
            Object collision=cls("dtb$a").getMethod("k",cls("dcc"),cls("jd")).invoke(state,world,origin);
            for(int face=0;face<6;face++)if((Boolean)cls("dfy").getMethod("a",cls("exv"),cls("ji")).invoke(null,collision,directions[face]))flags|=1<<(6+face);
            Object bottom=cls("exv").getMethod("a",cls("ji")).invoke(collision,directions[0]);
            String[] patterns={"F","G","H","I","J"};
            for(int bit=0;bit<patterns.length;bit++){var field=cls("dox").getDeclaredField(patterns[bit]);field.setAccessible(true);if((Boolean)wallCoverage.invoke(null,bottom,field.get(null)))flags|=1<<(12+bit);}
            Object info=cls("ent$c").getConstructor(cls("jd"),cls("dtc"),cls("ub")).newInstance(origin,state,null);
            Object converted=cls("enq").getMethod("a",cls("dcz"),cls("jd"),cls("jd"),cls("ent$c"),cls("ent$c"),cls("enp")).invoke(cls("ems").getField("b").get(null),null,origin,origin,info,info,settings);
            int target=(Integer)cls("dfy").getMethod("i",cls("dtc")).invoke(null,cls("ent$c").getMethod("b").invoke(converted));
            states.add(new int[]{flags,target});
        }
        System.out.println(new Gson().toJson(Map.of("version",3955,"states",states)));
    }
}
