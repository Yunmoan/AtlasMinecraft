/** Minimal player shell for the official ArmorStand.interactAt oracle.
 * Allocated without a constructor: only the held stack, mode and Level are
 * replaced. All armor-stand selection and swap logic is the original JAR. */
import com.mojang.authlib.GameProfile;
public final class AtlasArmorStandPlayer extends cmx {
 public dcw level; public cuq held; public boolean creative, spectator;
 public AtlasArmorStandPlayer(dcw level) { super(level,new jd(0,0,0),0,new GameProfile(new java.util.UUID(0,1),"Probe")); }
 @Override public boolean R_(){return spectator;}
 @Override public boolean f(){return creative;}
 @Override public boolean fL(){return creative;}
 @Override public dcw dO(){return level;}
 @Override public cuq b(bqq hand){return held;}
 @Override public void a(bqq hand,cuq stack){held=stack;}
}
