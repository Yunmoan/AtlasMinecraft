package atlasprobe;

import java.lang.instrument.*;
import java.security.ProtectionDomain;
import org.objectweb.asm.*;

/** Only fixes Entity's independently seeded RandomSource for reproducible probes.
 * UUID, item yaw, motion, dropping and NBT code all remain original methods. */
public final class EntityRandomProbeAgent {
    public static void premain(String ignored, Instrumentation instrumentation) {
        instrumentation.addTransformer(new ClassFileTransformer() {
            @Override public byte[] transform(ClassLoader loader,String name,Class<?> previous,ProtectionDomain domain,byte[] data) {
                if(!name.equals("bsr"))return null;
                var reader=new ClassReader(data);
                var writer=new ClassWriter(reader,ClassWriter.COMPUTE_MAXS);
                int[] replaced={0};
                reader.accept(new ClassVisitor(Opcodes.ASM9,writer) {
                    @Override public MethodVisitor visitMethod(int access,String name,String desc,String signature,String[] exceptions) {
                        var original=super.visitMethod(access,name,desc,signature,exceptions);
                        if(!name.equals("<init>"))return original;
                        return new MethodVisitor(Opcodes.ASM9,original) {
                            @Override public void visitMethodInsn(int opcode,String owner,String name,String desc,boolean itf) {
                                if(owner.equals("ayw")&&name.equals("a")&&desc.equals("()Layw;")) {
                                    super.visitLdcInsn("atlas.entitySeed");super.visitLdcInsn(0L);
                                    super.visitMethodInsn(Opcodes.INVOKESTATIC,"java/lang/Long","getLong","(Ljava/lang/String;J)Ljava/lang/Long;",false);
                                    super.visitMethodInsn(Opcodes.INVOKEVIRTUAL,"java/lang/Long","longValue","()J",false);
                                    super.visitMethodInsn(opcode,owner,name,"(J)Layw;",itf);replaced[0]++;
                                }else super.visitMethodInsn(opcode,owner,name,desc,itf);
                            }
                        };
                    }
                },0);
                if(replaced[0]!=1)throw new IllegalStateException("Expected one native entity RNG initialization: "+replaced[0]);
                System.setProperty("atlas.entitySeedApplied","true");
                return writer.toByteArray();
            }
        });
    }
}
