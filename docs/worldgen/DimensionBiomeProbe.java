package atlasprobe;

import java.io.*;
import java.lang.reflect.*;
import java.util.*;
import java.util.function.Function;

/** Actual dimension biome sources backed by the original RandomState. */
public final class DimensionBiomeProbe {
    public static void main(String[] args) throws Exception {
        Object version=Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a",Class.forName("ae")).invoke(null,version);
        Class.forName("akt").getMethod("a").invoke(null);
        Object registries=Class.forName("pd").getMethod("a").invoke(null);
        Object lookup=Class.forName("jo$a").getMethod("b").invoke(registries);
        Object biomes=Class.forName("jn$a").getMethod("b",Class.forName("akq")).invoke(lookup,Class.forName("lu").getField("aF").get(null));
        Object end=Class.forName("den").getMethod("a",Class.forName("jn")).invoke(null,biomes);
        Object preset=Class.forName("dek$a").getField("a").get(null);
        Object provider=Class.forName("dek$a").getMethod("c").invoke(preset);
        Method apply=Class.forName("dek$a$a").getMethod("apply",Function.class);apply.setAccessible(true);
        Object points=apply.invoke(provider,(Function<Object,Object>)key->{
            try { return Class.forName("jn").getMethod("b",Class.forName("akq")).invoke(biomes,key); }
            catch(ReflectiveOperationException e) { throw new RuntimeException(e); }
        });
        Object nether=Class.forName("dej").getMethod("a",Class.forName("def$c")).invoke(null,points);
        Method at=Class.forName("dea").getMethod("getNoiseBiome",int.class,int.class,int.class,Class.forName("def$f"));
        List<int[]> positions=new ArrayList<>();
        int[] borders={-7500000,-4097,-4096,-257,-256,-1,0,1,255,256,4096,4097,7500000};
        for(int x:borders)for(int z:borders)positions.add(new int[]{x,0,z});
        Random random=new Random(767);
        for(int i=0;i<512;i++)positions.add(new int[]{random.nextInt(15000001)-7500000,random.nextInt(129)-16,random.nextInt(15000001)-7500000});
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        for(String dimension:new String[]{"nether","end"})for(long seed:new long[]{0,1,-1,1234,9212937586943976949L,Long.MIN_VALUE}) {
            Object key=Class.forName("dzd").getField(dimension.equals("nether")?"f":"g").get(null);
            Object state=Class.forName("dzm").getMethod("a",Class.forName("jn$a"),Class.forName("akq"),long.class).invoke(null,lookup,key,seed);
            Object sampler=Class.forName("dzm").getMethod("b").invoke(state);
            for(int[] p:positions) {
                if(dimension.equals("nether")) {
                    Object target=Class.forName("def$f").getMethod("a",int.class,int.class,int.class).invoke(sampler,p[0],p[1],p[2]);
                    System.out.printf("dimension_climate %s %d %d %d %d",dimension,seed,p[0],p[1],p[2]);
                    for(String field:new String[]{"b","c","d","e","f","g"})System.out.print(" "+Class.forName("def$h").getMethod(field).invoke(target));
                    System.out.println();
                }
                Object biome=at.invoke(dimension.equals("nether")?nether:end,p[0],p[1],p[2],sampler);
                Object biomeKey=Class.forName("jm$c").getMethod("h").invoke(biome);
                System.out.printf("dimension_biome %s %d %d %d %d %s%n",dimension,seed,p[0],p[1],p[2],Class.forName("akq").getMethod("a").invoke(biomeKey));
            }
        }
    }
}
