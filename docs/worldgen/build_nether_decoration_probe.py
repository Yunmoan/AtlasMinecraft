"""Build an independent actual-JAR oracle for Nether decoration features."""
from pathlib import Path
import os
import subprocess

root = Path(__file__).resolve().parents[2]
source = (root / 'docs/worldgen/NetherForestProbe.java').read_text()
source = source.replace('NetherForestProbe', 'NetherDecorationProbe')
source = source.replace('Nether forest vegetation and huge fungi', 'Nether vines, glowstone and geological decorations')
begin = source.index('for(String name:new String[]{')
end = source.index('}) {', begin)
source = source[:begin] + 'for(String name:new String[]{"air","netherrack","warped_nylium","nether_wart_block","warped_wart_block","bedrock","water","lava","stone","basalt","blackstone","magma_block","soul_sand","nether_bricks","nether_brick_fence","nether_brick_stairs","nether_wart","chest","spawner","glowstone"' + source[end:]
begin = source.index('String[] names=')
end = source.index('Object world=Proxy',begin)
source = source[:begin] + '''String[] names={"twisting_vines","twisting_vines_bonemeal","weeping_vines","glowstone_extra","basalt_pillar","small_basalt_columns","large_basalt_columns","delta","basalt_blobs","blackstone_blobs"};
        for(String name:names)for(long seed:new long[]{0,1,-1,42,123456789,987654321,2,3})for(int scenario=0;scenario<28;scenario++) {
            final int mode=scenario;
            int[] origin=mode==1?new int[]{-1,41,-16}:mode==2?new int[]{15,41,15}:mode==18?new int[]{8,1,8}:mode==19?new int[]{8,0,8}:mode==20?new int[]{8,255,8}:mode==21?new int[]{8,128,8}:mode==22?new int[]{8,32,8}:mode==23?new int[]{8,33,8}:mode==24?new int[]{8,30,8}:new int[]{8,41,8};
            Map<String,Object> changed=new HashMap<>();
            boolean ceiling=name.equals("weeping_vines")||name.equals("glowstone_extra")||name.equals("basalt_pillar");
            java.util.function.Function<int[],Object> read=p->{
                if(p[1]<0||p[1]>=256)return blocks.get("air");
                String key=p[0]+" "+p[1]+" "+p[2];if(changed.containsKey(key))return changed.get(key);
                int dx=p[0]-origin[0],dy=p[1]-origin[1],dz=p[2]-origin[2];
                if(mode==3)return blocks.get("air");
                if(mode==4)return blocks.get("stone");
                int floor=origin[1]-(ceiling?12:1),roof=origin[1]+1;
                if(mode==15)floor+=Math.floorMod(dx+dz,5)-2;
                if(mode==25)floor+=Math.floorMod(dx*dx+dz*dz,7)-3;
                String support=switch(mode){case 5->"stone";case 6->"basalt";case 7->"blackstone";case 8->"bedrock";case 9->"nether_wart_block";case 10->"warped_wart_block";case 11->"warped_nylium";case 12->"lava";default->"netherrack";};
                if(mode==17)support=new String[]{"bedrock","magma_block","soul_sand","nether_bricks","nether_brick_fence","nether_brick_stairs","nether_wart","chest","spawner","netherrack"}[Math.floorMod(dx+dz,10)];
                if(p[1]<=floor)return blocks.get(p[1]==floor?support:"netherrack");
                if(ceiling&&p[1]==roof&&!(mode==16&&(dx!=0||dz!=0)))return blocks.get(support);
                if(mode==13&&dy==0)return blocks.get("lava");
                if(mode==14&&dy==0)return blocks.get("water");
                if(mode==26&&dy>=-6&&dy<=4&&Math.floorMod(dx*3+dz,5)==0)return blocks.get("netherrack");
                if(mode==27&&dy==0&&dx==0&&dz==0)return blocks.get("glowstone");
                return blocks.get("air");
            };
            ''' + source[end:]
source = source.replace('atlas-nether-forest-world','atlas-nether-decoration-world').replace('nether_forest minecraft:','nether_decoration minecraft:')
path=root/'docs/worldgen/NetherDecorationProbe.java'
path.write_text(source)
target=root/'artifacts/worldgen-probes'
classpath=os.pathsep.join([str(target),str(root/'vanilla/server-inner.jar')]+[str(p) for p in (root/'vanilla/libraries').rglob('*.jar')])
subprocess.run(['javac','-proc:none','-cp',classpath,'-d',str(target),str(path)],check=True)
with (root/'docs/worldgen/nether_decoration_1.21.1.txt').open('w') as out:
    subprocess.run(['java','-cp',classpath,'atlasprobe.NetherDecorationProbe'],stdout=out,check=True)
