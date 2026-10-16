from pathlib import Path
import os
import subprocess

root=Path(__file__).resolve().parents[2]
source=(root/'docs/worldgen/NetherDecorationProbe.java').read_text().replace('NetherDecorationProbe','ScatteredOreProbe')
begin=source.index('String[] names=')
end=source.index('for(String name:names)',begin)
source=source[:begin]+'''for(int variant=0;variant<8;variant++) {
            int size=new int[]{0,1,8,16,64,64,64,8}[variant];
            float chance=new float[]{0,0,.25f,.5f,1,.5f,.5f,0}[variant];
            List<Object> targets=new ArrayList<>();
            Object base=Class.forName("dtb$a").getMethod("b").invoke(blocks.get(variant==6?"stone":"netherrack"));
            Object predicate=Class.forName("emv").getConstructor(Class.forName("dfy")).newInstance(base);
            targets.add(Class.forName("eeq").getMethod("a",Class.forName("enn"),state).invoke(null,predicate,blocks.get(variant==7?"netherrack":"basalt")));
            if(variant==5||variant==6) {
                base=Class.forName("dtb$a").getMethod("b").invoke(blocks.get("netherrack"));
                predicate=Class.forName("emv").getConstructor(Class.forName("dfy")).newInstance(base);
                targets.add(Class.forName("eeq").getMethod("a",Class.forName("enn"),state).invoke(null,predicate,blocks.get("blackstone")));
            }
            Object config=Class.forName("eeq").getConstructor(List.class,int.class,float.class).newInstance(targets,size,chance);
            features.put("minecraft:scatter_custom_"+variant,Class.forName("ebq").getConstructor(Class.forName("ece"),Class.forName("eei")).newInstance(Class.forName("ece").getField("af").get(null),config));
        }
        String[] names={"ore_ancient_debris_large","ore_ancient_debris_small","scatter_custom_0","scatter_custom_1","scatter_custom_2","scatter_custom_3","scatter_custom_4","scatter_custom_5","scatter_custom_6","scatter_custom_7"};
        '''+source[end:]
source=source.replace('{"nylium","aM"}', '{"nylium","aM"},{"base_stone_nether","bi"}')
source=source.replace('nether_decoration minecraft:','scattered_ore minecraft:')
path=root/'docs/worldgen/ScatteredOreProbe.java';path.write_text(source)
target=root/'artifacts/worldgen-probes'
classpath=os.pathsep.join([str(target),str(root/'vanilla/server-inner.jar')]+[str(p) for p in (root/'vanilla/libraries').rglob('*.jar')])
subprocess.run(['javac','-proc:none','-cp',classpath,'-d',str(target),str(path)],check=True)
with(root/'docs/worldgen/scattered_ore_1.21.1.txt').open('w')as out:
    subprocess.run(['java','-cp',classpath,'atlasprobe.ScatteredOreProbe'],stdout=out,check=True)
