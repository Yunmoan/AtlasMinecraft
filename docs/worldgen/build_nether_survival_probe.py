from pathlib import Path
import os
import subprocess

root=Path(__file__).resolve().parents[2]
source=(root/'docs/worldgen/NetherForestProbe.java').read_text().replace('NetherForestProbe','NetherSurvivalProbe')
source=source.replace('{"nylium","aM"}', '{"nylium","aM"},{"soul_fire_base_blocks","aU"}')
begin=source.index('Object registries=')
source=source[:begin]+'''Class<?> pos=Class.forName("jd"),state=Class.forName("dtc");
        Object origin=pos.getConstructor(int.class,int.class,int.class).newInstance(-1,41,15);
        Method stateID=Class.forName("dfy").getMethod("i",state);
        Object fire=Class.forName("dga").getField("cr").get(null);
        Method canBurn=Class.forName("diy").getDeclaredMethod("f",state);canBurn.setAccessible(true);
        byte[] flags=new byte[(26684+7)/8];
        for(int id=0;id<26684;id++)if((boolean)canBurn.invoke(fire,Class.forName("dfy").getMethod("a",int.class).invoke(null,id)))flags[id/8]|=1<<(id%8);
        System.out.printf("fire_burnable_states 26684 %s%n",Base64.getEncoder().encodeToString(flags));
        String[] grounds={"dirt","grass_block","podzol","coarse_dirt","mycelium","rooted_dirt","moss_block","mud","muddy_mangrove_roots","farmland","clay","stone","sand","water","air","oak_leaves","mangrove_leaves","dirt_path","snow_block","crimson_nylium","warped_nylium","soul_sand","soul_soil","netherrack","oak_planks","crimson_planks","glass","oak_slab","oak_stairs","magma_block","bedrock","nether_wart_block","warped_wart_block","red_carpet","white_wool","short_grass","warped_roots","snow","lava"};
        String[] sides={"air","oak_planks","oak_leaves","crimson_planks","warped_wart_block","white_wool","red_carpet","oak_slab","oak_stairs","short_grass","warped_roots","stone","water"};
        String[] names={"fire","soul_fire","crimson_roots","warped_roots","crimson_fungus","warped_fungus","nether_sprouts"};
        Map<String,Object> blocks=new HashMap<>();List<String> all=new ArrayList<>(List.of(grounds));all.addAll(List.of(sides));all.addAll(List.of(names));
        for(String name:all){Object block=Class.forName("jz").getMethod("a",Class.forName("akr")).invoke(blockRegistry,Class.forName("akr").getMethod("a",String.class).invoke(null,"minecraft:"+name));blocks.put(name,Class.forName("dfy").getMethod("o").invoke(block));}
        for(String name:names)for(String ground:grounds)for(String side:sides)for(int variant=0;variant<2;variant++) {
            Object neighbor=blocks.get(side),below=blocks.get(ground);
            if(variant==1) {
                Object waterlogged=Class.forName("dts").getField("C").get(null);
                if((boolean)Class.forName("dte").getMethod("b",Class.forName("duf")).invoke(neighbor,waterlogged))neighbor=Class.forName("dte").getMethod("a",Class.forName("duf"),Comparable.class).invoke(neighbor,waterlogged,true);
                if((boolean)Class.forName("dte").getMethod("b",Class.forName("duf")).invoke(below,waterlogged))below=Class.forName("dte").getMethod("a",Class.forName("duf"),Comparable.class).invoke(below,waterlogged,true);
            }
            final Object neighborState=neighbor,groundState=below;
            Object world=Proxy.newProxyInstance(Class.forName("dcz").getClassLoader(),new Class<?>[]{Class.forName("dcz")},(proxy,method,values)->{
                if(method.getName().equals("a_")){int xx=(int)pos.getMethod("u").invoke(values[0]),yy=(int)pos.getMethod("v").invoke(values[0]),zz=(int)pos.getMethod("w").invoke(values[0]);if(xx==-1&&yy==40&&zz==15)return groundState;if(xx==0&&yy==41&&zz==15)return neighborState;return blocks.get("air");}
                throw new UnsupportedOperationException(method.toString());
            });
            boolean survives=(boolean)Class.forName("dtb$a").getMethod("a",Class.forName("dcz"),pos).invoke(blocks.get(name),world,origin);
            System.out.printf("nether_survival minecraft:%s %d %d %d%n",name,stateID.invoke(null,below),stateID.invoke(null,neighbor),survives?1:0);
        }
    }
}
'''
path=root/'docs/worldgen/NetherSurvivalProbe.java';path.write_text(source)
target=root/'artifacts/worldgen-probes'
classpath=os.pathsep.join([str(target),str(root/'vanilla/server-inner.jar')]+[str(p) for p in (root/'vanilla/libraries').rglob('*.jar')])
subprocess.run(['javac','-proc:none','-cp',classpath,'-d',str(target),str(path)],check=True)
with(root/'docs/worldgen/nether_survival_1.21.1.txt').open('w')as out:
    subprocess.run(['java','-cp',classpath,'atlasprobe.NetherSurvivalProbe'],stdout=out,check=True)
flags=(root/'docs/worldgen/nether_survival_1.21.1.txt').read_text().splitlines()[0]
(root/'internal/worldgen/fire_burnable_1.21.1.txt').write_text(flags+'\n')
