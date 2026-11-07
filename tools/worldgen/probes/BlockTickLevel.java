import java.util.*;

/** Defined ServerLevel for original deterministic scheduled block handlers. */
public class BlockTickLevel extends aqu {
    public Map<jd,dtc> blocks;
    public List<String> events;
    public List<int[]> inputs;
    public ayw random;
    public BlockTickLevel(){super(null,null,null,null,null,null,null,false,0L,null,false,null);}
    @Override public dtc a_(jd p){return blocks.getOrDefault(p,dga.a.o());}
    @Override public void a(jd p,dfy block,int delay){events.add("block_tick "+p.u()+" "+p.v()+" "+p.w()+" "+lt.e.b(block)+" "+delay+" 0");}
    @Override public void a(jd p,dfy block,int delay,eyx priority){events.add("block_tick "+p.u()+" "+p.v()+" "+p.w()+" "+lt.e.b(block)+" "+delay+" "+priority.a());}
    @Override public boolean a(jd p,dtc s,int flags){
        if(a_(p)==s)return false;blocks.put(p.j(),s);events.add("set "+p.u()+" "+p.v()+" "+p.w()+" "+dfy.i(s)+" "+flags);return true;
    }
    public static String handler(int id)throws Exception {
        for(Class<?> c=dfy.a(id).b().getClass();c!=null;c=c.getSuperclass())try {return c.getDeclaredMethod("a",dtc.class,aqu.class,jd.class,ayw.class).getDeclaringClass().getName();}catch(NoSuchMethodException e){}
        throw new IllegalStateException("missing scheduled tick handler");
    }
    public static int leafDistance(int id){return dki.n(dfy.a(id)).orElse(7);}
    public static String blockName(int id){return lt.e.b(dfy.a(id).b()).toString();}
    public List<Integer> postProcessStates(int neighbor){
        blocks=new LinkedHashMap<>();events=new ArrayList<>();jd p=new jd(8,64,8);
        for(ji face:ji.values())blocks.put(p.a(face),dfy.a(neighbor));
        var marked=new ArrayList<Integer>();
        for(int id=0;id<26684;id++){blocks.put(p,dfy.a(id));if(dfy.a(id).q(this,p))marked.add(id);}
        return marked;
    }
    private static dtc state(String name){return lt.e.a(akr.a("minecraft:"+name)).o();}
    public static int[] bubbleStates(){return new int[]{dfy.i(state("magma_block")),dfy.i(state("soul_sand")),dfy.i(state("bubble_column")),dfy.i((dtc)state("bubble_column").a(dge.b,false))};}
    public void configureBubble(int id,int scene,long seed){
        blocks=new LinkedHashMap<>();events=new ArrayList<>();random=new dyz(seed);
        jd p=new jd(8,64,8);dtc s=dfy.a(id);blocks.put(p,s);
        jd start=s.b() instanceof dge?p:p.d();
        jd obstacle=start.equals(p)?p.d():start;
        if(s.b() instanceof dge)blocks.put(p.e(),state(new String[]{"magma_block","soul_sand","bubble_column","stone","air"}[scene%5]));
        for(int y=start.v();y<start.v()+8;y++)if(y!=p.v())blocks.put(new jd(8,y,8),state("water"));
        if(scene==5)blocks.put(obstacle,(dtc)state("water").a(dts.aP,1));
        if(scene==6)blocks.put(obstacle,state("kelp"));
        if(scene==7)blocks.put(start.b(0,3,0),(dtc)state("water").a(dts.aP,1));
        if(scene==8)blocks.put(start.b(0,3,0),state("stone"));
        if(scene==9)blocks.put(start.b(0,3,0),(dtc)state("oak_slab").a(dts.C,true));
        if(scene==10||scene==11)for(int y=start.v();y<start.v()+8;y++)blocks.put(new jd(8,y,8),(dtc)state("bubble_column").a(dge.b,scene==10));
        if(scene==12)blocks.put(obstacle,state("air"));
        inputs=new ArrayList<>();for(var e:blocks.entrySet())inputs.add(new int[]{e.getKey().u(),e.getKey().v(),e.getKey().w(),dfy.i(e.getValue())});
    }
    public long place(){jd p=new jd(8,64,8);a_(p).a(this,p,state("air"),false);return random.g();}
    public void configure(int id,int scene,long seed){
        blocks=new LinkedHashMap<>();events=new ArrayList<>();random=new dyz(seed);
        jd p=new jd(8,64,8);blocks.put(p,dfy.a(id));
        if(scene>=1 && scene<=6)blocks.put(p.a(ji.values()[scene-1]),state("oak_log"));
        if(scene>=7 && scene<=12)blocks.put(p.a(ji.values()[scene-7]),(dtc)state("oak_leaves").a(dki.d,2));
        if(scene==13)for(ji face:ji.values())blocks.put(p.a(face),(dtc)state("oak_leaves").a(dki.d,face.ordinal()+1));
        if(scene==14){dtc s=state("scaffolding");blocks.put(p.e(),(dtc)s.a((duc)s.b().l().a("distance"),0));}
        if(scene==15){blocks.put(p.i(),state("oak_log"));blocks.put(p.e(),(dtc)state("oak_leaves").a(dki.d,1));}
        if(scene==16)for(ji face:ji.values())blocks.put(p.a(face),dga.b.o());
        inputs=new ArrayList<>();for(var e:blocks.entrySet())inputs.add(new int[]{e.getKey().u(),e.getKey().v(),e.getKey().w(),dfy.i(e.getValue())});
    }
    public long tick(){jd p=new jd(8,64,8);a_(p).a(this,p,random);return random.g();}
}
