from pathlib import Path
import os
import subprocess

root=Path(__file__).resolve().parents[2]
source=(root/'docs/worldgen/NetherDecorationProbe.java').read_text().replace('NetherDecorationProbe','NetherPatchProbe')
begin=source.index('String[] names=');end=source.index(';',begin)
source=source[:begin]+'String[] names={"patch_fire","patch_soul_fire","patch_crimson_roots"}'+source[end:]
source=source.replace('"glowstone"}', '"glowstone","soul_soil","crimson_roots"}')
source=source.replace('scenario<28','scenario<30').replace('case 12->"lava";','case 12->"lava";case 28->"soul_soil";case 29->"warped_nylium";')
source=source.replace('{"nylium","aM"}', '{"nylium","aM"},{"soul_fire_base_blocks","aU"}')
needle='boolean placed=(boolean)Class.forName("ebq")'
begin=source.index(needle);end=source.index('\n            List<String> output=',begin)
source=source[:begin]+'''int[] counts={0,0};
            if(args.length==0){counts[0]=1;if((boolean)Class.forName("ebq").getMethod("a",Class.forName("dds"),Class.forName("duz"),Class.forName("ayw"),pos).invoke(features.get("minecraft:"+name),world,generator,random,start))counts[1]++;}
            else {
                Object context=Class.forName("eiw").getConstructor(Class.forName("dds"),Class.forName("duz"),Optional.class).newInstance(world,generator,Optional.empty());
                Method positionsMethod=Class.forName("eiy").getMethod("a_",Class.forName("eiw"),Class.forName("ayw"),pos);
                Stream<?> positions=Stream.of(start);
                for(Object modifier:(List<?>)Class.forName("eiv").getMethod("c").invoke(placedFeatures.get("minecraft:"+name))) {
                    if(modifier.getClass().getName().equals("eii"))continue;
                    positions=positions.flatMap(p->{try{return(Stream<?>)positionsMethod.invoke(modifier,context,random,p);}catch(Exception e){throw new RuntimeException(e);}});
                }
                positions.forEach(p->{try{counts[0]++;if((boolean)Class.forName("ebq").getMethod("a",Class.forName("dds"),Class.forName("duz"),Class.forName("ayw"),pos).invoke(features.get("minecraft:"+name),world,generator,random,p))counts[1]++;}catch(Exception e){throw new RuntimeException(e);}});
            }
            boolean placed=counts[1]>0;'''+source[end:]
source=source.replace('nether_decoration minecraft:','nether_patch minecraft:')
source=source.replace('for(long seed:new long[]{0,1,-1,42,123456789,987654321,2,3})','for(long seed:(args.length==0?new long[]{0,1,-1,42,123456789,987654321,2,3}:java.util.stream.LongStream.range(0,64).toArray()))')
needle='System.out.printf("nether_patch minecraft:'
begin=source.index(needle)
source=source[:begin]+'if(args.length>0)System.out.printf("nether_patch_placed minecraft:%s %d %d %d %d %d %d %d %d %d %s%n",name,seed,scenario,origin[0],origin[1],origin[2],counts[0],counts[1],changed.size(),Class.forName("dzx").getMethod("g").invoke(random),hash);else '+source[begin:]
path=root/'docs/worldgen/NetherPatchProbe.java';path.write_text(source)
target=root/'artifacts/worldgen-probes'
classpath=os.pathsep.join([str(target),str(root/'vanilla/server-inner.jar')]+[str(p) for p in (root/'vanilla/libraries').rglob('*.jar')])
subprocess.run(['javac','-proc:none','-cp',classpath,'-d',str(target),str(path)],check=True)
with(root/'docs/worldgen/nether_patch_1.21.1.txt').open('w')as out:
    subprocess.run(['java','-cp',classpath,'atlasprobe.NetherPatchProbe'],stdout=out,check=True)
with(root/'docs/worldgen/nether_patch_placed_1.21.1.txt').open('w')as out:
    subprocess.run(['java','-cp',classpath,'atlasprobe.NetherPatchProbe','placed'],stdout=out,check=True)
