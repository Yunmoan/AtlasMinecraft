from pathlib import Path
import os
import subprocess

root=Path(__file__).resolve().parents[2]
source=(root/'docs/worldgen/NetherDecorationProbe.java').read_text().replace('NetherDecorationProbe','EndFeatureProbe')
source=source.replace('getField("f").get(null));','getField("g").get(null));')
source=source.replace('biomes.get("minecraft:nether_wastes")','biomes.get("minecraft:the_end")')
begin=source.index('for(String name:new String[]{');end=source.index('}) {',begin)
source=source[:begin]+'for(String name:new String[]{"air","end_stone","stone","water","lava","obsidian","bedrock","iron_bars","end_gateway","chorus_plant","chorus_flower"'+source[end:]
begin=source.index('String[] names=');end=source.index('Object world=Proxy',begin)
source=source[:begin]+'''Object gatewayNone=Class.forName("eeh").getMethod("a").invoke(null);
        Object gatewayExit=pos.getConstructor(int.class,int.class,int.class).newInstance(-100,50,123);
        Object gatewayInexact=Class.forName("eeh").getMethod("a",pos,boolean.class).invoke(null,gatewayExit,false);
        var configured=Class.forName("ebq").getConstructor(Class.forName("ece"),Class.forName("eei"));
        features.put("minecraft:end_gateway_none",configured.newInstance(Class.forName("ece").getField("M").get(null),gatewayNone));
        features.put("minecraft:end_gateway_inexact",configured.newInstance(Class.forName("ece").getField("M").get(null),gatewayInexact));
        Object guarded=Class.forName("edo$a").getConstructor(int.class,int.class,int.class,int.class,boolean.class).newInstance(8,8,2,76,true);
        Object small=Class.forName("edo$a").getConstructor(int.class,int.class,int.class,int.class,boolean.class).newInstance(10,10,3,80,false);
        Object beam=pos.getConstructor(int.class,int.class,int.class).newInstance(0,128,0);
        Object guardConfig=Class.forName("efc").getConstructor(boolean.class,List.class,pos).newInstance(false,List.of(guarded),null);
        Object beamConfig=Class.forName("efc").getConstructor(boolean.class,List.class,pos).newInstance(true,List.of(guarded,small),beam);
        features.put("minecraft:end_spike_guarded",configured.newInstance(Class.forName("ece").getField("K").get(null),guardConfig));
        features.put("minecraft:end_spike_beam",configured.newInstance(Class.forName("ece").getField("K").get(null),beamConfig));
        Field unsafeField=sun.misc.Unsafe.class.getDeclaredField("theUnsafe");unsafeField.setAccessible(true);
        var unsafe=(sun.misc.Unsafe)unsafeField.get(null);
        Object server=unsafe.allocateInstance(Class.forName("apn")),level=unsafe.allocateInstance(Class.forName("aqu"));
        Object worldData=Proxy.newProxyInstance(Class.forName("erl").getClassLoader(),new Class<?>[]{Class.forName("erl")},(proxy,method,values)->{if(method.getName().equals("K"))return Class.forName("cpn").getField("g").get(null);throw new UnsupportedOperationException(method.toString());});
        Field serverData=Class.forName("net.minecraft.server.MinecraftServer").getDeclaredField("j");serverData.setAccessible(true);serverData.set(server,worldData);
        Field serverField=Class.forName("aqu").getDeclaredField("J");serverField.setAccessible(true);serverField.set(level,server);
        long[] seeds={0,1,-1,42,123456789,987654321,2,3};
        Map<Long,List<?>> layouts=new HashMap<>();
        for(long seed:seeds){Object seedWorld=Proxy.newProxyInstance(Class.forName("dds").getClassLoader(),new Class<?>[]{Class.forName("dds")},(proxy,method,values)->{if(method.getName().equals("C"))return seed;throw new UnsupportedOperationException(method.toString());});List<?> layout=(List<?>)Class.forName("edo").getMethod("a",Class.forName("dds")).invoke(null,seedWorld);layouts.put(seed,layout);for(int i=0;i<layout.size();i++){Object s=layout.get(i);System.out.printf("end_spike_layout %d %d %s %s %s %s %s%n",seed,i,Class.forName("edo$a").getMethod("a").invoke(s),Class.forName("edo$a").getMethod("b").invoke(s),Class.forName("edo$a").getMethod("c").invoke(s),Class.forName("edo$a").getMethod("d").invoke(s),(boolean)Class.forName("edo$a").getMethod("e").invoke(s)?1:0);}}
        String[] names={"end_island","chorus_plant","end_gateway_return","end_platform","end_spike","end_gateway_none","end_gateway_inexact","end_spike_guarded","end_spike_beam"};
        for(String name:names)for(long seed:seeds)for(int scenario=0;scenario<18;scenario++) {
            final int mode=scenario;
            int[] origin=mode==1?new int[]{-1,70,-16}:mode==2?new int[]{15,70,15}:mode==12?new int[]{8,1,8}:mode==13?new int[]{8,0,8}:mode==14?new int[]{8,255,8}:mode==15?new int[]{8,254,8}:new int[]{8,70,8};
            if(name.equals("end_spike")&&mode<10){Object s=layouts.get(seed).get(mode);origin[0]=(int)Class.forName("edo$a").getMethod("a").invoke(s);origin[2]=(int)Class.forName("edo$a").getMethod("b").invoke(s);}
            Map<String,Object> changed=new HashMap<>(),entities=new HashMap<>();
            List<Object> crystals=new ArrayList<>();
            java.util.function.Function<int[],Object> read=p->{
                if(p[1]<0||p[1]>=256)return blocks.get("air");
                String key=p[0]+" "+p[1]+" "+p[2];if(changed.containsKey(key))return changed.get(key);
                int dx=p[0]-origin[0],dy=p[1]-origin[1],dz=p[2]-origin[2];
                if(mode==3)return blocks.get("air");
                if(mode==4)return blocks.get("stone");
                int floor=origin[1]-1;if(name.startsWith("end_spike"))floor=66;
                if(mode==10)floor+=Math.floorMod(dx+dz,5)-2;
                if(p[1]<=floor)return blocks.get(mode==5?"stone":mode==6?"water":mode==7?"obsidian":"end_stone");
                if(mode==8&&dy==0)return blocks.get("water");
                if(mode==9&&dy==0)return blocks.get("lava");
                if(mode==11&&dy>=2&&dy<=5&&Math.abs(dx)<=2&&Math.abs(dz)<=2)return blocks.get("stone");
                if(mode==16&&dy>=0&&dy<=6&&dx==2)return blocks.get("chorus_plant");
                if(mode==17&&dy>=0&&dy<=6&&Math.floorMod(dx+dz,3)==0)return blocks.get("end_stone");
                return blocks.get("air");
            };
            '''+source[end:]
source=source.replace('if(m.equals("toString"))','if(m.equals("E"))return level;if(m.equals("C"))return seed;if(m.equals("b")&&values.length==1&&Class.forName("bsr").isInstance(values[0])){crystals.add(values[0]);return true;}\n                if(m.equals("toString"))')
source=source.replace('if(m.equals("a_"))return s;', '''if(m.equals("c_"))return entities.get(p[0]+" "+p[1]+" "+p[2]);
                    if(m.equals("a_"))return s;''')
source=source.replace('changed.put(p[0]+" "+p[1]+" "+p[2],m.equals("b")?blocks.get("air"):values[1]);return true;', '''changed.put(p[0]+" "+p[1]+" "+p[2],m.equals("b")?blocks.get("air"):values[1]);
                        if(m.equals("a")&&values[1]==blocks.get("end_gateway"))entities.put(p[0]+" "+p[1]+" "+p[2],Class.forName("dry").getConstructor(pos,state).newInstance(values[0],values[1]));
                        return true;''')
source=source.replace('nether_decoration minecraft:','end_feature minecraft:')
needle='System.out.printf("end_feature minecraft:'
begin=source.index(needle);end=source.index('\n        }',begin)
source=source[:end]+'''
            for(var entry:entities.entrySet()) {
                Object saved=Class.forName("dqh").getMethod("b",Class.forName("jo$a")).invoke(entry.getValue(),registries);
                var bytes=io.netty.buffer.Unpooled.buffer();Object buffer=Class.forName("vw").getConstructor(io.netty.buffer.ByteBuf.class).newInstance(bytes);Class.forName("vw").getMethod("a",Class.forName("uy")).invoke(buffer,saved);byte[] data=new byte[bytes.readableBytes()];bytes.readBytes(data);bytes.release();
                System.out.printf("end_gateway_nbt minecraft:%s %d %d %s %s%n",name,seed,scenario,entry.getKey(),HexFormat.of().formatHex(data));
            }
            for(Object crystal:crystals) {
                Object target=Class.forName("chz").getMethod("p").invoke(crystal);String beamText=target==null?"none":x.invoke(target)+","+y.invoke(target)+","+z.invoke(target);
                System.out.printf("end_crystal minecraft:%s %d %d %s %s %s %s %s %s%n",name,seed,scenario,Class.forName("bsr").getMethod("dt").invoke(crystal),Class.forName("bsr").getMethod("dv").invoke(crystal),Class.forName("bsr").getMethod("dz").invoke(crystal),Float.floatToRawIntBits((float)Class.forName("bsr").getMethod("dE").invoke(crystal)),(boolean)Class.forName("bsr").getMethod("cv").invoke(crystal)?1:0,beamText);
            }
'''+source[end:]
path=root/'docs/worldgen/EndFeatureProbe.java';path.write_text(source)
target=root/'artifacts/worldgen-probes'
classpath=os.pathsep.join([str(target),str(root/'vanilla/server-inner.jar')]+[str(p) for p in (root/'vanilla/libraries').rglob('*.jar')])
subprocess.run(['javac','-proc:none','-cp',classpath,'-d',str(target),str(path)],check=True)
with(root/'docs/worldgen/end_features_1.21.1.txt').open('w')as out:
    subprocess.run(['java','-cp',classpath,'atlasprobe.EndFeatureProbe'],stdout=out,check=True)
