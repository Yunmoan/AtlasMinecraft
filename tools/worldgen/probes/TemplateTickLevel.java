import java.util.*;

/** Authored LevelAccessor for original template-edge and neighbor-shape code.
 * Writes are raw, as in WorldGenRegion, with no ServerLevel onPlace recursion. */
public class TemplateTickLevel extends aqu {
    public Map<jd,dtc> blocks;
    public List<Map<String,Object>> requests;
    public dtc fallback;
    public dwy dimension;
    public ayw random;
    public TemplateTickLevel(){super(null,null,null,null,null,null,null,false,0L,null,false,null);}
    @Override public int I_(){return -64;}
    @Override public int am(){return 320;}
    @Override public dwy D_(){return dimension;}
    @Override public cpl J(){return cpn.g;}
    @Override public ayw E_(){return random;}
    @Override public dtc a_(jd p){return blocks.getOrDefault(p,fallback);}
    @Override public epe b_(jd p){return a_(p).u();}
    @Override public boolean a(jd p,dtc state,int flags){if(a_(p)==state)return false;blocks.put(p.j(),state);return true;}
    private void request(jd p,String kind,String name,int delay){
        var r=new LinkedHashMap<String,Object>();r.put("kind",kind);r.put("type",name);r.put("x",p.u());r.put("y",p.v());r.put("z",p.w());r.put("delay",delay);r.put("priority",0);requests.add(r);
    }
    @Override public void a(jd p,dfy block,int delay){request(p,"block",lt.e.b(block).toString(),delay);}
    @Override public void a(jd p,epd fluid,int delay){request(p,"fluid",lt.c.b(fluid).toString(),delay);}
    public static int[] states(){
        var ids=new ArrayList<Integer>();
        for(int id=0;id<26684;id++){
            dfy b=dfy.a(id).b();String name=lt.e.b(b).toString();
            if(b instanceof diu||b instanceof dgd||b instanceof dki||b instanceof dko||Set.of("minecraft:seagrass","minecraft:tall_seagrass","minecraft:kelp","minecraft:kelp_plant","minecraft:magma_block","minecraft:soul_sand","minecraft:bubble_column").contains(name))ids.add(id);
        }
        return ids.stream().mapToInt(Integer::intValue).toArray();
    }
    public Map<String,Object> run(int id,int scene,int shape,boolean finish)throws Exception {
        blocks=new LinkedHashMap<>();requests=new ArrayList<>();
        random=new dyz(1234L);
        String[] fills={"air","stone","oak_log","oak_leaves","water","sand","magma_block","soul_sand","bubble_column","water","lava","oak_leaves"};
        fallback=lt.e.a(akr.a("minecraft:"+fills[scene])).o();
        if(scene==9||scene==10)fallback=fallback.a(dko.b,1);
        if(scene==11)fallback=fallback.a(dki.f,true);
        var placed=new ArrayList<jd>();var mask=new exf(3,3,3);
        if(shape==0)placed.add(new jd(-1,60,15));
        else if(shape==1){placed.add(new jd(-1,60,15));placed.add(new jd(0,60,15));placed.add(new jd(0,60,16));}
        else for(int x=0;x<3;x++)for(int y=0;y<3;y++)for(int z=0;z<3;z++)if(x==0||x==2||y==0||y==2||z==0||z==2)placed.add(new jd(x-1,y+60,z+15));
        for(jd p:placed){blocks.put(p,dfy.a(id));mask.c(p.u()+1,p.v()-60,p.w()-15);}
        ent.a(this,2,mask,-1,60,15);
        if(finish)for(jd p:placed){dtc old=a_(p),updated=dfy.b(old,(dcx)this,p);if(old!=updated)a(p,updated,2);}
        var result=new LinkedHashMap<String,Object>();result.put("state",id);result.put("fill",dfy.i(fallback));result.put("shape",shape);result.put("finish",finish);
        var positions=new ArrayList<int[]>();for(var e:blocks.entrySet()){jd p=e.getKey();positions.add(new int[]{p.u(),p.v(),p.w(),dfy.i(e.getValue())});}
        positions.sort((a,b)->{for(int i=0;i<3;i++){int c=Integer.compare(a[i],b[i]);if(c!=0)return c;}return 0;});
        result.put("blocks",positions);result.put("requests",requests);result.put("next",Long.toUnsignedString(random.g()));return result;
    }
}
