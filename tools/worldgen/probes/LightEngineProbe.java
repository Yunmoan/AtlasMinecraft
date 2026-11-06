package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.security.MessageDigest;
import java.util.*;
import java.util.function.BiConsumer;

/** Executes the actual vanilla light engines against isolated three-by-three chunks. */
public final class LightEngineProbe {
    static Class<?> pos, state, getter, chunk;
    static Method x,y,z,decode,emission;
    static Map<String,Object> blocks=new HashMap<>();
    static Object air,bedrock;
    static int scenario;
    static Object at(int px,int py,int pz) {
        if ((px>>4)<-1 || (px>>4)>1 || (pz>>4)<-1 || (pz>>4)>1) return bedrock;
        if (py<0 || py>=32) return air;
        if (py==0) return blocks.get("stone");
        if (py==12 && (scenario!=1 || px!=8 || pz!=8)) return blocks.get("stone");
        if (scenario!=3 && px==15 && py==4 && pz==8) return blocks.get("torch");
        if (px==8 && py==4 && pz==8) return blocks.get("glowstone");
        if (px==10 && py==4 && pz==8) return blocks.get("stone_slab");
        if (scenario==2 && px>=5 && px<=11 && pz>=5 && pz<=11 && py>=13 && py<=15) return blocks.get("water");
        return air;
    }
    static Object position(int px,int py,int pz) throws Exception { return pos.getConstructor(int.class,int.class,int.class).newInstance(px,py,pz); }
    static String key(int cx,int cz) { return cx+","+cz; }
    public static void main(String[] args) throws Exception {
        Object version=Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a",Class.forName("ae")).invoke(null,version);
        Class.forName("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        pos=Class.forName("jd");state=Class.forName("dtc");getter=Class.forName("dcc");chunk=Class.forName("dvk");
        x=pos.getMethod("u");y=pos.getMethod("v");z=pos.getMethod("w");
        emission=Class.forName("dtb$a").getMethod("h");
        Object registry=Class.forName("lt").getField("e").get(null);
        for(String name:new String[]{"air","bedrock","stone","torch","glowstone","stone_slab","water"}) {
            Object block=Class.forName("jz").getMethod("a",Class.forName("akr")).invoke(registry,Class.forName("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name));
            blocks.put(name,Class.forName("dfy").getMethod("o").invoke(block));
        }
        air=blocks.get("air");bedrock=blocks.get("bedrock");
        for(scenario=0;scenario<4;scenario++) {
            Object world=Proxy.newProxyInstance(getter.getClassLoader(),new Class[]{getter},(p,m,v)-> {
                if(m.getName().equals("I_")) return 0;
                if(m.getName().equals("J_")) return 32;
                if(m.getName().equals("a_")) return at((int)x.invoke(v[0]),(int)y.invoke(v[0]),(int)z.invoke(v[0]));
                if(m.getName().equals("b_")) return Class.forName("dtb$a").getMethod("u").invoke(at((int)x.invoke(v[0]),(int)y.invoke(v[0]),(int)z.invoke(v[0])));
                if(m.getName().equals("c_")) return null;
                if(m.isDefault()) return InvocationHandler.invokeDefault(p,m,v);
                throw new UnsupportedOperationException(m.toString());
            });
            Map<String,Object> chunks=new HashMap<>();
            Method blocked=Class.forName("eoo").getDeclaredMethod("a",getter,pos,state,pos,state); blocked.setAccessible(true);
            Method setSource=Class.forName("eoo").getDeclaredMethod("b",int.class,int.class);setSource.setAccessible(true);
            for(int cx=-1;cx<=1;cx++) for(int cz=-1;cz<=1;cz++) {
                int chunkX=cx,chunkZ=cz;
                Object sources=Class.forName("eoo").getConstructor(Class.forName("dcy")).newInstance(world);
                for(int lz=0;lz<16;lz++) for(int lx=0;lx<16;lx++) {
                    int px=cx*16+lx,pz=cz*16+lz,lowest=-1;
                    for(int py=31;py>=0;py--) if((boolean)blocked.invoke(null,world,position(px,py+1,pz),at(px,py+1,pz),position(px,py,pz),at(px,py,pz))) { lowest=py+1;break; }
                    setSource.invoke(sources,lz*16+lx,lowest);
                }
                Object proxy=Proxy.newProxyInstance(chunk.getClassLoader(),new Class[]{chunk},(p,m,v)-> {
                    if(m.getName().equals("B")) return sources;
                    if(m.getName().equals("a") && v[0] instanceof BiConsumer) {
                        BiConsumer<Object,Object> consumer=(BiConsumer<Object,Object>)v[0];
                        for(int py=0;py<32;py++) for(int lz=0;lz<16;lz++) for(int lx=0;lx<16;lx++) {
                            Object s=at(chunkX*16+lx,py,chunkZ*16+lz);
                            if((int)emission.invoke(s)>0) consumer.accept(position(chunkX*16+lx,py,chunkZ*16+lz),s);
                        }
                        return null;
                    }
                    return getter.getMethod(m.getName(),m.getParameterTypes()).invoke(world,v);
                });
                chunks.put(key(cx,cz),proxy);
            }
            Object provider=Proxy.newProxyInstance(Class.forName("dvl").getClassLoader(),new Class[]{Class.forName("dvl")},(p,m,v)-> {
                if(m.getName().equals("q")) return world;
                if(m.getName().equals("c")) return chunks.get(key((int)v[0],(int)v[1]));
                if(m.getReturnType()==void.class) return null;
                throw new UnsupportedOperationException(m.toString());
            });
            Class<?> engineType=Class.forName("eot"), section=Class.forName("kf"), chunkPos=Class.forName("dcd");
            Object engine=engineType.getConstructor(Class.forName("dvl"),boolean.class,boolean.class).newInstance(provider,true,true);
            for(int cx=-1;cx<=1;cx++) for(int cz=-1;cz<=1;cz++) {
                for(int sy=0;sy<2;sy++) engineType.getMethod("a",section,boolean.class).invoke(engine,section.getMethod("a",int.class,int.class,int.class).invoke(null,cx,sy,cz),false);
                Object cp=chunkPos.getConstructor(int.class,int.class).newInstance(cx,cz);
                engineType.getMethod("a",chunkPos,boolean.class).invoke(engine,cp,true);
                engineType.getMethod("b",chunkPos).invoke(engine,cp);
            }
            for(int i=0;(boolean)engineType.getMethod("K_").invoke(engine);i++) {
                if(i>100) throw new IllegalStateException("light engine did not settle");
                engineType.getMethod("a").invoke(engine);
            }
            Object[] layers=Class.forName("ddf").getEnumConstants();
            for(int kind=0;kind<2;kind++) {
                Object layer=engineType.getMethod("a",Class.forName("ddf")).invoke(engine,layers[kind]);
                byte[] result=new byte[32*256];
                for(int py=0;py<32;py++) for(int pz=0;pz<16;pz++) for(int px=0;px<16;px++)
                    result[py*256+pz*16+px]=(byte)(int)Class.forName("eor").getMethod("b",pos).invoke(layer,position(px,py,pz));
                System.out.printf("light_case %d %s %s%n",scenario,layers[kind],HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(result)));
            }
        }
    }
}
