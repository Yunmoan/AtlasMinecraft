package atlasprobe;

import java.io.FileDescriptor;
import java.io.FileOutputStream;
import java.io.PrintStream;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HexFormat;
import java.util.List;

/** Calls StructureTemplate.processBlockInfos with the actual complete portal
 * processor chain, before clipping, state transforms and template placement. */
public final class RuinedPortalProcessorsProbe {
    static Class<?> cls(String name)throws Exception{return Class.forName(name);}
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        RuinedPortalProbeSupport.bindTags();Object registries=IglooProbeSupport.registryAccess(),manager=IglooProbeSupport.manager(RuinedPortalProbeSupport.names());
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        for(String name:RuinedPortalProbeSupport.names())for(int rotation=0;rotation<4;rotation++)for(int mirror:new int[]{0,2})for(int variant=0;variant<6;variant++)for(int mode=0;mode<6;mode++)for(int[] anchor:new int[][]{{0,64,0},{-16,60,-16},{15,66,15}}) {
            Object template=cls("enu").getMethod("a",cls("akr")).invoke(manager,cls("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name));
            Object size=cls("ent").getMethod("a").invoke(template),pivot=cls("jd").getConstructor(int.class,int.class,int.class).newInstance((Integer)cls("kh").getMethod("u").invoke(size)/2,0,(Integer)cls("kh").getMethod("w").invoke(size)/2);
            Object origin=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(anchor[0],anchor[1],anchor[2]);
            var makeSettings=cls("emf").getDeclaredMethod("a",cls("dkv"),cls("dmm"),cls("emf$b"),cls("jd"),cls("emf$a"));makeSettings.setAccessible(true);
            Object placement=cls("emf$b").getEnumConstants()[variant==2?2:variant==4?5:0];
            Object settings=makeSettings.invoke(null,cls("dkv").getEnumConstants()[mirror],cls("dmm").getEnumConstants()[rotation],placement,pivot,RuinedPortalProbeSupport.properties(variant));
            var palettes=cls("ent").getDeclaredField("l");palettes.setAccessible(true);Object palette=((List<?>)palettes.get(template)).get(0);
            List<?> blocks=(List<?>)cls("ent$a").getMethod("a").invoke(palette);
            Object world=RuinedPortalProbeSupport.world(mode,new HashMap<>(),new HashMap<>(),registries);
            List<?> processed=(List<?>)cls("ent").getMethod("a",cls("ddl"),cls("jd"),cls("jd"),cls("enp"),List.class).invoke(null,world,origin,origin,settings,blocks);
            List<String> rows=new ArrayList<>();for(Object info:processed){Object p=cls("ent$c").getMethod("a").invoke(info),state=cls("ent$c").getMethod("b").invoke(info),data=cls("ent$c").getMethod("c").invoke(info);rows.add(cls("jd").getMethod("u").invoke(p)+" "+cls("jd").getMethod("v").invoke(p)+" "+cls("jd").getMethod("w").invoke(p)+" "+cls("dfy").getMethod("i",cls("dtc")).invoke(null,state)+" "+(data==null?"null":IglooProbeSupport.canonicalNBT(data)));}
            // Preserve the actual processor return order, not a position sort.
            String hash=HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(String.join("\n",rows).getBytes(StandardCharsets.UTF_8)));
            System.out.printf("processors %s %d %d %d %d %d %d %d %d %s%n",name,rotation,mirror,variant,mode,anchor[0],anchor[1],anchor[2],rows.size(),hash);
        }
    }
}
