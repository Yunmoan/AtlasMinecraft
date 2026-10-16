package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.function.BiConsumer;

/** Real 1.21.1 checkBlock updates; includes every block and light sample. */
public final class LightUpdateProbe {
    static Class<?> pos, state, getter, chunk, engineType;
    static Method x, y, z, emission, stateID;
    static Object air, bedrock;
    static final Map<String,Object> blocks = new HashMap<>();
    static final Map<String,Object[]> columns = new LinkedHashMap<>();
    static final Map<String,Object> sources = new HashMap<>();
    static String key(int cx, int cz) { return cx + "," + cz; }
    static Object position(int px, int py, int pz) throws Exception {
        return pos.getConstructor(int.class,int.class,int.class).newInstance(px,py,pz);
    }
    static Object at(int px, int py, int pz) {
        Object[] column = columns.get(key(px>>4,pz>>4));
        if (column == null) return bedrock;
        if (py < 0 || py >= 32) return air;
        return column[py*256+(pz&15)*16+(px&15)];
    }
    static Object property(Object value, String name, String text) throws Exception {
        for (Object p : (Collection<?>)Class.forName("dte").getMethod("B").invoke(value)) {
            if (!Class.forName("duf").getMethod("f").invoke(p).equals(name)) continue;
            Object parsed = ((Optional<?>)Class.forName("duf").getMethod("b",String.class).invoke(p,text)).orElseThrow();
            return Class.forName("dte").getMethod("a",Class.forName("duf"),Comparable.class).invoke(value,p,parsed);
        }
        throw new IllegalArgumentException("missing property " + name);
    }
    static void settle(Object engine) throws Exception {
        for (int i=0;(boolean)engineType.getMethod("K_").invoke(engine);i++) {
            if (i>100) throw new IllegalStateException("light update did not settle");
            engineType.getMethod("a").invoke(engine);
        }
    }
    static void output(Object engine, int step, boolean sky) throws Exception {
        Object[] kinds = Class.forName("ddf").getEnumConstants();
        for (int kind=0;kind<2;kind++) {
            Object layer=engineType.getMethod("a",Class.forName("ddf")).invoke(engine,kinds[kind]);
            for(int cx=-1;cx<=1;cx++) for(int cz=-1;cz<=1;cz++) {
                byte[] values=new byte[32*256];
                if (kind != 0 || sky) {
                    for(int py=0;py<32;py++) for(int lz=0;lz<16;lz++) for(int lx=0;lx<16;lx++)
                        values[py*256+lz*16+lx]=(byte)(int)Class.forName("eor").getMethod("b",pos).invoke(layer,position(cx*16+lx,py,cz*16+lz));
                }
                System.out.printf("light_update %d %d %d %s %s%n",step,cx,cz,kinds[kind],HexFormat.of().formatHex(values));
            }
        }
    }
    static void run(String name, boolean sky, boolean roof, int px, int py, int pz, Object... edits) throws Exception {
        columns.clear(); sources.clear();
        System.out.printf("light_update_case %s %s%n",name,sky);
        for(int cx=-1;cx<=1;cx++) for(int cz=-1;cz<=1;cz++) {
            Object[] values=new Object[32*256]; Arrays.fill(values,air);
            for(int i=0;i<256;i++) values[i]=blocks.get("stone");
            if (roof) for(int i=12*256;i<13*256;i++) values[i]=blocks.get("stone");
            columns.put(key(cx,cz),values);
        }
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
        Method setSource=Class.forName("eoo").getDeclaredMethod("b",int.class,int.class); setSource.setAccessible(true);
        for(int cx=-1;cx<=1;cx++) for(int cz=-1;cz<=1;cz++) {
            int chunkX=cx,chunkZ=cz;
            Object source=Class.forName("eoo").getConstructor(Class.forName("dcy")).newInstance(world);
            sources.put(key(cx,cz),source);
            for(int lz=0;lz<16;lz++) for(int lx=0;lx<16;lx++) {
                int bx=cx*16+lx,bz=cz*16+lz,lowest=-1;
                for(int by=31;by>=0;by--) if((boolean)blocked.invoke(null,world,position(bx,by+1,bz),at(bx,by+1,bz),position(bx,by,bz),at(bx,by,bz))) {lowest=by+1;break;}
                setSource.invoke(source,lz*16+lx,lowest);
            }
            chunks.put(key(cx,cz),Proxy.newProxyInstance(chunk.getClassLoader(),new Class[]{chunk},(p,m,v)-> {
                if(m.getName().equals("B")) return source;
                if(m.getName().equals("a") && v[0] instanceof BiConsumer) {
                    BiConsumer<Object,Object> consumer=(BiConsumer<Object,Object>)v[0];
                    for(int by=0;by<32;by++) for(int lz=0;lz<16;lz++) for(int lx=0;lx<16;lx++) {
                        Object s=at(chunkX*16+lx,by,chunkZ*16+lz);
                        if((int)emission.invoke(s)>0) consumer.accept(position(chunkX*16+lx,by,chunkZ*16+lz),s);
                    }
                    return null;
                }
                return getter.getMethod(m.getName(),m.getParameterTypes()).invoke(world,v);
            }));
            byte[] states=new byte[32*256*2];
            Object[] values=columns.get(key(cx,cz));
            for(int i=0;i<values.length;i++) {
                int id=(int)stateID.invoke(null,values[i]); states[2*i]=(byte)(id>>8);states[2*i+1]=(byte)id;
            }
            System.out.printf("light_update_states %d %d %s%n",cx,cz,HexFormat.of().formatHex(states));
        }
        Object provider=Proxy.newProxyInstance(Class.forName("dvl").getClassLoader(),new Class[]{Class.forName("dvl")},(p,m,v)-> {
            if(m.getName().equals("q")) return world;
            if(m.getName().equals("c")) return chunks.get(key((int)v[0],(int)v[1]));
            if(m.getReturnType()==void.class) return null;
            throw new UnsupportedOperationException(m.toString());
        });
        Class<?> section=Class.forName("kf"),chunkPos=Class.forName("dcd");
        Object engine=engineType.getConstructor(Class.forName("dvl"),boolean.class,boolean.class).newInstance(provider,true,sky);
        for(int cx=-1;cx<=1;cx++) for(int cz=-1;cz<=1;cz++) {
            for(int sy=0;sy<2;sy++) engineType.getMethod("a",section,boolean.class).invoke(engine,section.getMethod("a",int.class,int.class,int.class).invoke(null,cx,sy,cz),false);
            Object cp=chunkPos.getConstructor(int.class,int.class).newInstance(cx,cz);
            engineType.getMethod("a",chunkPos,boolean.class).invoke(engine,cp,true);
            engineType.getMethod("b",chunkPos).invoke(engine,cp);
        }
        settle(engine); output(engine,0,sky);
        int step=0;
        for(Object value:edits) {
            columns.get(key(px>>4,pz>>4))[py*256+(pz&15)*16+(px&15)]=value;
            // Same source-height maintenance used by LevelChunk.setBlockState.
            Class.forName("eoo").getMethod("a",getter,int.class,int.class,int.class).invoke(sources.get(key(px>>4,pz>>4)),world,px,py,pz);
            engineType.getMethod("a",pos).invoke(engine,position(px,py,pz));
            settle(engine);
            System.out.printf("light_update_edit %d %d %d %d %d%n",++step,px,py,pz,(int)stateID.invoke(null,value));
            output(engine,step,sky);
        }
    }
    public static void main(String[] args) throws Exception {
        Object version=Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a",Class.forName("ae")).invoke(null,version);
        Class.forName("akt").getMethod("a").invoke(null);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        pos=Class.forName("jd");state=Class.forName("dtc");getter=Class.forName("dcc");chunk=Class.forName("dvk");engineType=Class.forName("eot");
        x=pos.getMethod("u");y=pos.getMethod("v");z=pos.getMethod("w");
        emission=Class.forName("dtb$a").getMethod("h");stateID=Class.forName("dfy").getMethod("i",state);
        Object registry=Class.forName("lt").getField("e").get(null);
        for(String name:new String[]{"air","bedrock","stone","dirt","torch","glowstone","oak_slab","oak_stairs","water","glass"}) {
            Object block=Class.forName("jz").getMethod("a",Class.forName("akr")).invoke(registry,Class.forName("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name));
            blocks.put(name,Class.forName("dfy").getMethod("o").invoke(block));
        }
        air=blocks.get("air");bedrock=blocks.get("bedrock");
        run("roof",true,true,8,12,8,air,blocks.get("stone"));
        run("border_glowstone",true,false,15,4,8,blocks.get("glowstone"),air);
        run("negative_border_torch",true,false,-1,4,8,blocks.get("torch"),air);
        run("slab_faces",true,false,8,8,8,property(blocks.get("oak_slab"),"type","bottom"),property(blocks.get("oak_slab"),"type","top"),air);
        run("water_glass",true,false,8,8,8,blocks.get("water"),blocks.get("glass"),air);
        Object stair=property(blocks.get("oak_stairs"),"half","top");
        run("stairs",true,false,8,8,8,stair,property(property(stair,"facing","east"),"shape","outer_right"),air);
        run("no_sky_emitter",false,false,15,4,8,blocks.get("glowstone"),air);
        run("equivalent_opaque",true,false,8,0,8,blocks.get("dirt"),blocks.get("stone"));
        run("equivalent_transparent",true,false,8,8,8,blocks.get("glass"),air);
    }
}
