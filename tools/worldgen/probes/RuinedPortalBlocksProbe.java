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
import java.util.Map;

/** Actual complete RuinedPortalPiece postProcess, including netherrack spread,
 * drip columns, vines/leaves, liquid preservation, chest NBT and RNG. */
public final class RuinedPortalBlocksProbe {
    static Class<?> cls(String n)throws Exception{return Class.forName(n);}
    static String digest(List<String> rows)throws Exception{rows.sort(String::compareTo);return HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(String.join("\n",rows).getBytes(StandardCharsets.UTF_8)));}
    public static void main(String[] args)throws Exception {
        Object version=cls("t").getMethod("a").invoke(null);cls("ab").getMethod("a",cls("ae")).invoke(null,version);cls("akt").getMethod("a").invoke(null);
        RuinedPortalProbeSupport.bindTags();Object registries=IglooProbeSupport.registryAccess(),manager=IglooProbeSupport.manager(RuinedPortalProbeSupport.names());
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        for(String name:RuinedPortalProbeSupport.names())for(int rotation=0;rotation<4;rotation++)for(int mirror:new int[]{0,2})for(int variant=0;variant<6;variant++)for(int mode=0;mode<6;mode++)for(int[] anchor:new int[][]{{0,64,0},{-16,60,-16},{15,66,15}})for(int outside=0;outside<2;outside++) {
            long seed=new long[]{0,1,-1,1234}[(rotation+mirror+variant+mode+anchor[0])&3];
            Object key=cls("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name),template=cls("enu").getMethod("a",cls("akr")).invoke(manager,key);
            Object size=cls("ent").getMethod("a").invoke(template),pivot=cls("jd").getConstructor(int.class,int.class,int.class).newInstance((Integer)cls("kh").getMethod("u").invoke(size)/2,0,(Integer)cls("kh").getMethod("w").invoke(size)/2);
            Object origin=cls("jd").getConstructor(int.class,int.class,int.class).newInstance(anchor[0],anchor[1],anchor[2]);
            Object placement=cls("emf$b").getEnumConstants()[variant==2?2:variant==4?5:0];
            Object piece=cls("emf").getConstructor(cls("enu"),cls("jd"),cls("emf$b"),cls("emf$a"),cls("akr"),cls("ent"),cls("dmm"),cls("dkv"),cls("jd")).newInstance(manager,origin,placement,RuinedPortalProbeSupport.properties(variant),key,template,cls("dmm").getEnumConstants()[rotation],cls("dkv").getEnumConstants()[mirror],pivot);
            Object box=cls("ejv").getMethod("f").invoke(piece),center=cls("ejj").getMethod("g").invoke(box);
            int chunkX=((Integer)cls("jd").getMethod("u").invoke(center)>>4)+outside,chunkZ=(Integer)cls("jd").getMethod("w").invoke(center)>>4;
            Object clip=cls("ejj").getConstructor(int.class,int.class,int.class,int.class,int.class,int.class).newInstance(chunkX*16,-64,chunkZ*16,chunkX*16+15,319,chunkZ*16+15),chunk=cls("dcd").getConstructor(int.class,int.class).newInstance(chunkX,chunkZ);
            Map<String,Object> changed=new HashMap<>(),entities=new HashMap<>();Object world=RuinedPortalProbeSupport.world(mode,changed,entities,registries);
            Object random=cls("dzx").getConstructor(cls("ayw")).newInstance(cls("dzz").getConstructor(long.class).newInstance(seed));
            cls("ejv").getMethod("a",cls("dds"),cls("ddq"),cls("duz"),cls("ayw"),cls("ejj"),cls("dcd"),cls("jd")).invoke(piece,world,null,null,random,clip,chunk,origin);
            List<String> edits=new ArrayList<>(),fullEntities=new ArrayList<>();for(var entry:changed.entrySet())edits.add(entry.getKey()+" "+cls("dfy").getMethod("i",cls("dtc")).invoke(null,entry.getValue()));
            for(var entry:entities.entrySet()){Object tag=cls("dqh").getMethod("b",cls("jo$a")).invoke(entry.getValue(),registries);fullEntities.add(entry.getKey()+" "+IglooProbeSupport.canonicalNBT(tag));}
            Object saved=cls("ejv").getMethod("a",cls("ekh")).invoke(piece,(Object)null);
            String debugKey=name+" "+seed+" "+rotation+" "+mirror+" "+variant+" "+mode+" "+anchor[0]+" "+anchor[1]+" "+anchor[2]+" "+outside;
            if("1".equals(System.getenv("ATLAS_RUINED_PORTAL_DEBUG"))||debugKey.equals(System.getenv("ATLAS_RUINED_PORTAL_DEBUG"))){edits.sort(String::compareTo);if(System.getenv("ATLAS_RUINED_PORTAL_DEBUG_OUTPUT")!=null)java.nio.file.Files.writeString(java.nio.file.Path.of(System.getenv("ATLAS_RUINED_PORTAL_DEBUG_OUTPUT")),String.join("\n",edits));System.err.println(saved);throw new IllegalStateException("explicit ruined portal debug stop");}
            System.out.printf("portal %s "+"%d ".repeat(12)+"%s %d %s %d %s%n",name,seed,rotation,mirror,variant,mode,anchor[0],anchor[1],anchor[2],chunkX,chunkZ,outside,edits.size(),digest(edits),entities.size(),digest(fullEntities),cls("ayw").getMethod("g").invoke(random),digest(new ArrayList<>(List.of(IglooProbeSupport.canonicalNBT(saved)))));
        }
    }
}
